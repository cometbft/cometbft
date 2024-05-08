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
	go func(oracleInfo *types.OracleInfo) {
		for {
			select {
			case <-oracleInfo.StopChannel:
				return
			default:
				interval := oracleInfo.Config.SignInterval
				if interval == 0 {
					interval = 100 * time.Millisecond
				}
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
		case vote := <-oracleInfo.SignVotesChan:
			votes = append(votes, vote)
			continue
		default:
		}
		break
	}

	if len(votes) == 0 {
		return
	}

	// batch sign the new votes, along with existing votes in gossipVoteBuffer, if any
	validatorIndex, _ := consensusState.Validators.GetByAddress(oracleInfo.PubKey.Address())
	if validatorIndex == -1 {
		log.Errorf("processSignVoteQueue: unable to find validator index")
		return
	}

	// append new batch into unsignedVotesBuffer, need to mutex lock as it will clash with concurrent pruning
	oracleInfo.UnsignedVoteBuffer.UpdateMtx.Lock()
	oracleInfo.UnsignedVoteBuffer.Buffer = append(oracleInfo.UnsignedVoteBuffer.Buffer, votes...)

	unsignedVotes := []*oracleproto.Vote{}

	for _, vote := range oracleInfo.UnsignedVoteBuffer.Buffer {
		unsignedVotes = append(unsignedVotes, vote)
	}

	// sort the votes so that we can rebuild it in a deterministic order, when uncompressing
	SortOracleVotes(unsignedVotes)

	// batch sign the entire unsignedVoteBuffer and add to gossipBuffer
	newGossipVote := &oracleproto.GossipedVotes{
		Validator:       oracleInfo.PubKey.Address(),
		ValidatorIndex:  validatorIndex,
		SignedTimestamp: time.Now().Unix(),
		Votes:           unsignedVotes,
	}

	oracleInfo.UnsignedVoteBuffer.UpdateMtx.Unlock()

	// signing of vote should append the signature field of gossipVote
	if err := oracleInfo.PrivValidator.SignOracleVote("", newGossipVote); err != nil {
		log.Errorf("processSignVoteQueue: error signing oracle votes")
		return
	}

	// need to mutex lock as it will clash with concurrent gossip
	oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
	address := oracleInfo.PubKey.Address().String()
	oracleInfo.GossipVoteBuffer.Buffer[address] = newGossipVote
	oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
}

func PruneVoteBuffers(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	go func(oracleInfo *types.OracleInfo) {
		maxGossipVoteAge := oracleInfo.Config.MaxGossipVoteAge
		if maxGossipVoteAge == 0 {
			maxGossipVoteAge = 2
		}
		pruneInterval := oracleInfo.Config.PruneInterval
		if pruneInterval == 0 {
			if pruneInterval == 0 {
				pruneInterval = 500 * time.Millisecond
			}
		}

		ticker := time.Tick(pruneInterval)
		for range ticker {
			lastBlockTime := consensusState.GetState().LastBlockTime

			if !contains(oracleInfo.BlockTimestamps, lastBlockTime.Unix()) {
				oracleInfo.BlockTimestamps = append(oracleInfo.BlockTimestamps, lastBlockTime.Unix())
			}

			if len(oracleInfo.BlockTimestamps) < maxGossipVoteAge {
				continue
			}

			if len(oracleInfo.BlockTimestamps) > maxGossipVoteAge {
				oracleInfo.BlockTimestamps = oracleInfo.BlockTimestamps[1:]
			}

			oracleInfo.UnsignedVoteBuffer.UpdateMtx.Lock()
			// prune votes that are older than the maxGossipVoteAge (in terms of block height)
			newVotes := []*oracleproto.Vote{}
			unsignedVoteBuffer := oracleInfo.UnsignedVoteBuffer.Buffer
			for _, vote := range unsignedVoteBuffer {
				if vote.Timestamp >= oracleInfo.BlockTimestamps[0] {
					newVotes = append(newVotes, vote)
				}
			}
			oracleInfo.UnsignedVoteBuffer.Buffer = newVotes
			oracleInfo.UnsignedVoteBuffer.UpdateMtx.Unlock()

			oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
			gossipBuffer := oracleInfo.GossipVoteBuffer.Buffer

			// prune gossip votes that have not been updated in the last x amt of blocks, where x is the maxGossipVoteAge
			for valAddr, gossipVote := range gossipBuffer {
				if gossipVote.SignedTimestamp < oracleInfo.BlockTimestamps[0] {
					delete(gossipBuffer, valAddr)
				}
			}
			oracleInfo.GossipVoteBuffer.Buffer = gossipBuffer
			oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
		}
	}(oracleInfo)
}

func contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Run run oracles
func Run(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	log.Info("[oracle] Service started.")
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
