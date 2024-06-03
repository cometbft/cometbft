package runner

import (
	"context"
	"sort"

	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cometbft/cometbft/oracle/service/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cs "github.com/cometbft/cometbft/consensus"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
)

func RunProcessSignVoteQueue(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	interval := oracleInfo.Config.SignInterval
	if interval == 0 {
		interval = 100 * time.Millisecond
	}

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
	oracleInfo.UnsignedVoteBuffer.UpdateMtx.Lock()
	oracleInfo.UnsignedVoteBuffer.Buffer = append(oracleInfo.UnsignedVoteBuffer.Buffer, votes...)

	unsignedVotes := []*oracleproto.Vote{}
	unsignedVotes = append(unsignedVotes, oracleInfo.UnsignedVoteBuffer.Buffer...)

	oracleInfo.UnsignedVoteBuffer.UpdateMtx.Unlock()

	// sort the votes so that we can rebuild it in a deterministic order, when uncompressing
	SortOracleVotes(unsignedVotes)

	// batch sign the entire unsignedVoteBuffer and add to gossipBuffer
	newGossipVote := &oracleproto.GossipedVotes{
		Validator:       oracleInfo.PubKey.Address(),
		SignedTimestamp: time.Now().Unix(),
		Votes:           unsignedVotes,
	}

	// signing of vote should append the signature field of gossipVote
	if err := oracleInfo.PrivValidator.SignOracleVote("", newGossipVote); err != nil {
		log.Errorf("processSignVoteQueue: error signing oracle votes")
		return
	}

	// need to mutex lock as it will clash with concurrent gossip
	preLockTime := time.Now().UnixMicro()
	oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
	address := oracleInfo.PubKey.Address().String()
	oracleInfo.GossipVoteBuffer.Buffer[address] = newGossipVote
	oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
	postLockTime := time.Now().UnixMicro()
	diff := postLockTime - preLockTime
	if diff > 1000 {
		log.Infof("Updating gossip lock took %v microseconds", diff)
	}
}

func PruneVoteBuffers(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	go func(oracleInfo *types.OracleInfo) {
		// only keep votes that are less than 2 blocks old
		maxOracleGossipBlocksDelayed := oracleInfo.Config.MaxOracleGossipBlocksDelayed
		if maxOracleGossipBlocksDelayed == 0 {
			maxOracleGossipBlocksDelayed = 2
		}

		// only keep votes that are less than 30s old
		maxOracleGossipAge := oracleInfo.Config.MaxOracleGossipAge
		if maxOracleGossipAge == 0 {
			maxOracleGossipAge = 30
		}

		pruneInterval := oracleInfo.Config.PruneInterval
		if pruneInterval == 0 {
			pruneInterval = 500 * time.Millisecond
		}

		ticker := time.Tick(pruneInterval)
		for range ticker {
			lastBlockTime := consensusState.GetState().LastBlockTime.Unix()
			currTimestampsLen := len(oracleInfo.BlockTimestamps)

			if currTimestampsLen == 0 {
				oracleInfo.BlockTimestamps = append(oracleInfo.BlockTimestamps, lastBlockTime)
				continue
			}

			if oracleInfo.BlockTimestamps[currTimestampsLen-1] != lastBlockTime {
				oracleInfo.BlockTimestamps = append(oracleInfo.BlockTimestamps, lastBlockTime)
			}

			// if chain is stale and not enough blockTimestamps have been accumulated, add extra check to see if earliest block timestamp is older than the latest allowable timestamp
			latestAllowableTimestamp := time.Now().Unix() - int64(maxOracleGossipAge)
			if len(oracleInfo.BlockTimestamps) < maxOracleGossipBlocksDelayed && oracleInfo.BlockTimestamps[0] > latestAllowableTimestamp {
				continue
			}

			// only keep last x number of block timestamps, where x = maxOracleGossipBlocksDelayed
			if len(oracleInfo.BlockTimestamps) > maxOracleGossipBlocksDelayed {
				oracleInfo.BlockTimestamps = oracleInfo.BlockTimestamps[1:]
			}

			// prune votes that are older than the latestAllowableTimestamp, which is the max(earliest block timestamp collected, current time - maxOracleGossipAge)
			if oracleInfo.BlockTimestamps[0] > latestAllowableTimestamp {
				latestAllowableTimestamp = oracleInfo.BlockTimestamps[0]
			}

			oracleInfo.UnsignedVoteBuffer.UpdateMtx.Lock()
			newVotes := []*oracleproto.Vote{}
			unsignedVoteBuffer := oracleInfo.UnsignedVoteBuffer.Buffer
			for _, vote := range unsignedVoteBuffer {
				if vote.Timestamp >= latestAllowableTimestamp {
					newVotes = append(newVotes, vote)
				}
			}
			oracleInfo.UnsignedVoteBuffer.Buffer = newVotes
			oracleInfo.UnsignedVoteBuffer.UpdateMtx.Unlock()

			preLockTime := time.Now().UnixMicro()
			oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
			gossipBuffer := oracleInfo.GossipVoteBuffer.Buffer

			// prune gossipedVotes that are older than the latestAllowableTimestamp, which is the max(earliest block timestamp collected, current time - maxOracleGossipAge)
			for valAddr, gossipVote := range gossipBuffer {
				if gossipVote.SignedTimestamp < latestAllowableTimestamp {
					delete(gossipBuffer, valAddr)
				}
			}
			oracleInfo.GossipVoteBuffer.Buffer = gossipBuffer
			oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
			postLockTime := time.Now().UnixMicro()
			diff := postLockTime - preLockTime
			if diff > 1000 {
				log.Infof("Pruning gossip lock took %v microseconds", diff)
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
