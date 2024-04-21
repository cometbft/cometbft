package runner

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cometbft/cometbft/oracle/service/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
)

func RunProcessSignVoteQueue(oracleInfo *types.OracleInfo) {
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
				ProcessSignVoteQueue(oracleInfo)
			}
		}
	}(oracleInfo)
}

func ProcessSignVoteQueue(oracleInfo *types.OracleInfo) {
	for votes := range oracleInfo.SignVotesChan {
		if len(votes) == 0 {
			return
		}

		// new batch of unsigned votes
		newUnsignedVotes := &types.UnsignedVotes{
			Timestamp: uint64(time.Now().Unix()),
			Votes:     votes,
		}

		// append new batch into unsignedVotesBuffer, need to mutex lock as it will clash with concurrent pruning
		oracleInfo.UnsignedVoteBuffer.UpdateMtx.Lock()
		oracleInfo.UnsignedVoteBuffer.Buffer = append(oracleInfo.UnsignedVoteBuffer.Buffer, newUnsignedVotes)
		oracleInfo.UnsignedVoteBuffer.UpdateMtx.Unlock()

		// loop through unsignedVoteBuffer and combine all votes
		var batchVotes = []*oracleproto.Vote{}
		oracleInfo.UnsignedVoteBuffer.UpdateMtx.RLock()
		for _, unsignedVotes := range oracleInfo.UnsignedVoteBuffer.Buffer {
			batchVotes = append(batchVotes, unsignedVotes.Votes...)
		}
		oracleInfo.UnsignedVoteBuffer.UpdateMtx.RUnlock()

		// batch sign the entire unsignedVoteBuffer and add to gossipBuffer
		newGossipVote := &oracleproto.GossipVote{
			Validator: oracleInfo.PubKey.Address(),
			PublicKey: oracleInfo.PubKey.Bytes(),
			SignType:  oracleInfo.PubKey.Type(),
			Votes:     batchVotes,
		}

		// signing of vote should append the signature field and timestamp field of gossipVote
		if err := oracleInfo.PrivValidator.SignOracleVote("", newGossipVote); err != nil {
			log.Errorf("error signing oracle votes")
			continue
		}

		// replace current gossipVoteBuffer with new one
		address := oracleInfo.PubKey.Address().String()

		// need to mutex lock as it will clash with concurrent gossip
		oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
		oracleInfo.GossipVoteBuffer.Buffer[address] = newGossipVote
		log.Infof("adding new gossipBuffer at time: %v", newGossipVote.SignedTimestamp)
		oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
	}
}

func PruneGossipVoteBuffer(oracleInfo *types.OracleInfo) {
	go func(oracleInfo *types.OracleInfo) {
		interval := 60 * time.Second
		ticker := time.Tick(interval)
		for range ticker {
			oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
			currTime := uint64(time.Now().Unix())
			buffer := oracleInfo.GossipVoteBuffer.Buffer

			// prune gossip vote that have signed timestamps older than 60 secs
			for valAddr, gossipVote := range oracleInfo.GossipVoteBuffer.Buffer {
				if gossipVote.SignedTimestamp < currTime-uint64(interval) {
					log.Info(currTime - uint64(interval))
					log.Infof("DELETING STALE GOSSIP BUFFER (%v) FOR VAL: %s", gossipVote.SignedTimestamp, valAddr)
					delete(buffer, valAddr)
				}
			}
			oracleInfo.GossipVoteBuffer.Buffer = buffer
			oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
		}
	}(oracleInfo)
}

func PruneUnsignedVoteBuffer(oracleInfo *types.OracleInfo) {
	go func(oracleInfo *types.OracleInfo) {
		interval := oracleInfo.Config.PruneInterval
		if interval == 0 {
			interval = 4 * time.Second
		}
		ticker := time.Tick(interval)
		for range ticker {
			oracleInfo.UnsignedVoteBuffer.UpdateMtx.RLock()
			// prune everything older than 4 secs
			currTime := uint64(time.Now().Unix())
			numberOfVotesToPrune := 0
			count := 0
			unsignedVoteBuffer := oracleInfo.UnsignedVoteBuffer.Buffer
			for _, unsignedVotes := range unsignedVoteBuffer {
				// unsigned votes are arranged from least recent to most recent
				timestamp := unsignedVotes.Timestamp
				if timestamp <= currTime-uint64(interval) {
					numberOfVotesToPrune++
					count += len(unsignedVotes.Votes)
				} else {
					// everything beyond is more recent hence we can early terminate
					break
				}
			}
			oracleInfo.UnsignedVoteBuffer.UpdateMtx.RUnlock()

			if numberOfVotesToPrune > 0 {
				oracleInfo.UnsignedVoteBuffer.UpdateMtx.Lock()
				oracleInfo.UnsignedVoteBuffer.Buffer = oracleInfo.UnsignedVoteBuffer.Buffer[numberOfVotesToPrune:]
				oracleInfo.UnsignedVoteBuffer.UpdateMtx.Unlock()
			}
		}
	}(oracleInfo)
}

// Run run oracles
func Run(oracleInfo *types.OracleInfo) {
	log.Info("[oracle] Service started.")
	RunProcessSignVoteQueue(oracleInfo)
	PruneUnsignedVoteBuffer(oracleInfo)
	PruneGossipVoteBuffer(oracleInfo)
	// start to take votes from app
	for {
		res, err := oracleInfo.ProxyApp.PrepareOracleVotes(context.Background(), &abcitypes.RequestPrepareOracleVotes{})
		if err != nil {
			log.Error(err)
		}

		votes := []*oracleproto.Vote{}

		for _, vote := range res.Votes {
			newVote := oracleproto.Vote{
				Validator: oracleInfo.PubKey.Address(),
				OracleId:  vote.OracleId,
				Data:      vote.Data,
				Timestamp: uint64(vote.Timestamp),
			}
			votes = append(votes, &newVote)
		}

		log.Infof("RESULTS: %v", res.Votes)

		oracleInfo.SignVotesChan <- votes

		time.Sleep(1000 * time.Millisecond)
	}
}
