package runner

import (
	"context"
	"fmt"
	"sort"

	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/oracle/service/utils"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cs "github.com/cometbft/cometbft/consensus"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
)

func RunProcessSignVoteQueue(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	// sign votes every x milliseconds, where x = Config.SignInterval
	interval := oracleInfo.Config.SignInterval

	go func(oracleInfo *types.OracleInfo) {
		for {
			select {
			case <-oracleInfo.StopChannel:
				return
			default:
				time.Sleep(interval)
				ProcessSignVoteQueue(oracleInfo, consensusState)
			}
		}
	}(oracleInfo)
}

func ProcessSignVoteQueue(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	votes := []*oracleproto.Vote{}

	for {
		select {
		case newVote := <-oracleInfo.SignVotesChan:
			votes = append(votes, newVote)
			continue
		default:
		}
		break
	}

	if len(votes) == 0 {
		return
	}

	// batch sign the new votes, along with existing votes in gossipVoteBuffer, if any
	// append new batch into unsignedVotesBuffer, need to mutex lock as it will clash with concurrent pruning
	oracleInfo.UnsignedVoteBuffer.Lock()
	oracleInfo.UnsignedVoteBuffer.Buffer = append(oracleInfo.UnsignedVoteBuffer.Buffer, votes...)

	unsignedVotes := []*oracleproto.Vote{}
	unsignedVotes = append(unsignedVotes, oracleInfo.UnsignedVoteBuffer.Buffer...)

	oracleInfo.UnsignedVoteBuffer.Unlock()

	// sort the votes so that we can rebuild it in a deterministic order, when uncompressing
	SortOracleVotes(unsignedVotes)

	// batch sign the entire unsignedVoteBuffer and add to gossipBuffer
	newGossipVote := &oracleproto.GossipedVotes{
		PubKey:          oracleInfo.PubKey.Bytes(),
		SignedTimestamp: time.Now().Unix(),
		Votes:           unsignedVotes,
	}

	// set sigPrefix based on account type and sign type
	sigPrefix, err := utils.FormSignaturePrefix(oracleInfo.Config.EnableSubAccountSigning, oracleInfo.PubKey.Type())
	if err != nil {
		log.Errorf("processSignVoteQueue: unable to form sig prefix: %v", err)
		return
	}

	TIMEOUT := time.Second * 5
	chainState, err := consensusState.GetStateWithTimeout(TIMEOUT)
	if err != nil {
		log.Errorf("timed out trying to get chain state within %v", TIMEOUT)
		return
	}

	// signing of vote should append the signature field of gossipVote
	if err := oracleInfo.PrivValidator.SignOracleVote(chainState.ChainID, newGossipVote, sigPrefix); err != nil {
		log.Errorf("processSignVoteQueue: error signing oracle votes: %v", err)
		return
	}

	// need to mutex lock as it will clash with concurrent gossip
	preLockTime := time.Now().UnixMilli()
	oracleInfo.GossipVoteBuffer.Lock()
	address := oracleInfo.PubKey.Address().String()
	oracleInfo.GossipVoteBuffer.Buffer[address] = newGossipVote
	oracleInfo.GossipVoteBuffer.Unlock()
	postLockTime := time.Now().UnixMilli()
	diff := postLockTime - preLockTime
	if diff > 100 {
		log.Warnf("WARNING!!! Updating gossip lock took %v milliseconds", diff)
	}
}

func PruneVoteBuffers(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	go func(oracleInfo *types.OracleInfo) {
		// only keep votes that are less than x blocks old, where x = Config.MaxOracleGossipBlocksDelayed
		maxOracleGossipBlocksDelayed := oracleInfo.Config.MaxOracleGossipBlocksDelayed
		// only keep votes that are less than x seconds old, where x = Config.MaxOracleGossipAge
		maxOracleGossipAge := oracleInfo.Config.MaxOracleGossipAge
		// run pruner every x milliseconds, where x = Config.PruneInterval
		pruneInterval := oracleInfo.Config.PruneInterval

		ticker := time.Tick(pruneInterval)
		for range ticker {
			TIMEOUT := time.Second * 3
			chainState, err := consensusState.GetStateWithTimeout(TIMEOUT)
			if err != nil {
				log.Warnf("timed out trying to get chain state after %v", TIMEOUT)
			} else {
				lastBlockTime := chainState.LastBlockTime.Unix()
				currTimestampsLen := len(oracleInfo.BlockTimestamps)

				if currTimestampsLen == 0 {
					oracleInfo.BlockTimestamps = append(oracleInfo.BlockTimestamps, lastBlockTime)
					continue
				}

				if oracleInfo.BlockTimestamps[currTimestampsLen-1] != lastBlockTime {
					oracleInfo.BlockTimestamps = append(oracleInfo.BlockTimestamps, lastBlockTime)
				}
			}

			// only keep last x number of block timestamps, where x = maxOracleGossipBlocksDelayed
			if len(oracleInfo.BlockTimestamps) > maxOracleGossipBlocksDelayed {
				oracleInfo.BlockTimestamps = oracleInfo.BlockTimestamps[1:]
			}

			latestAllowableTimestamp := time.Now().Unix() - int64(maxOracleGossipAge)
			// prune votes that are older than the latestAllowableTimestamp, which is the max(earliest block timestamp collected, current time - maxOracleGossipAge)
			if len(oracleInfo.BlockTimestamps) == maxOracleGossipBlocksDelayed && oracleInfo.BlockTimestamps[0] > latestAllowableTimestamp {
				latestAllowableTimestamp = oracleInfo.BlockTimestamps[0]
			}

			oracleInfo.UnsignedVoteBuffer.Lock()
			newVotes := []*oracleproto.Vote{}
			unsignedVoteBuffer := oracleInfo.UnsignedVoteBuffer.Buffer
			visitedVoteMap := make(map[string]struct{})
			for _, vote := range unsignedVoteBuffer {
				// check for dup votes
				key := fmt.Sprintf("%v:%v", vote.Timestamp, vote.OracleId)
				_, exists := visitedVoteMap[key]
				if exists {
					continue
				}

				visitedVoteMap[key] = struct{}{}

				// also prune votes for a given oracle id and timestamp, that have already been committed as results on chain
				res, err := oracleInfo.ProxyApp.DoesOracleResultExist(context.Background(), &abcitypes.RequestDoesOracleResultExist{Key: key})
				if err != nil {
					log.Warnf("PruneVoteBuffers: unable to check if oracle result exist for vote: %v: %v", vote, err)
				}

				if res.DoesExist {
					continue
				}

				if vote.Timestamp >= latestAllowableTimestamp {
					newVotes = append(newVotes, vote)
				}
			}
			oracleInfo.UnsignedVoteBuffer.Buffer = newVotes
			oracleInfo.UnsignedVoteBuffer.Unlock()

			preLockTime := time.Now().UnixMilli()
			oracleInfo.GossipVoteBuffer.Lock()
			gossipBuffer := oracleInfo.GossipVoteBuffer.Buffer

			// prune gossipedVotes that are older than the latestAllowableTimestamp, which is the max(earliest block timestamp collected, current time - maxOracleGossipAge)
			for valAddr, gossipVote := range gossipBuffer {
				if gossipVote.SignedTimestamp < latestAllowableTimestamp {
					delete(gossipBuffer, valAddr)
				}
			}
			oracleInfo.GossipVoteBuffer.Buffer = gossipBuffer
			oracleInfo.GossipVoteBuffer.Unlock()
			postLockTime := time.Now().UnixMilli()
			diff := postLockTime - preLockTime
			if diff > 100 {
				log.Warnf("WARNING!!! Pruning gossip lock took %v milliseconds", diff)
			}
		}
	}(oracleInfo)
}

// Run run oracles
func Run(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	RunProcessSignVoteQueue(oracleInfo, consensusState)
	PruneVoteBuffers(oracleInfo, consensusState)
	// start to take votes from app
	for {
		res, err := oracleInfo.ProxyApp.FetchOracleVotes(context.Background(), &abcitypes.RequestFetchOracleVotes{})
		if err != nil {
			log.Errorf("app not ready: %v, retrying...", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if res.Vote == nil {
			continue
		}

		// drop old votes when channel hits half cap
		if len(oracleInfo.SignVotesChan) >= cap(oracleInfo.SignVotesChan)/2 {
			<-oracleInfo.SignVotesChan
		}

		oracleInfo.SignVotesChan <- res.Vote
	}
}

func SortOracleVotes(votes []*oracleproto.Vote) {
	sort.SliceStable(votes,
		func(i, j int) bool {
			if votes[i].Timestamp != votes[j].Timestamp {
				return votes[i].Timestamp < votes[j].Timestamp
			}
			if votes[i].OracleId != votes[j].OracleId {
				return votes[i].OracleId < votes[j].OracleId
			}
			return votes[i].Data < votes[j].Data
		})
}
