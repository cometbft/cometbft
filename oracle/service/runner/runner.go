package runner

import (
	"context"
	"sync"
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
		log.Errorf("unable to find validator index")
		return
	}

	newGossipVote := &oracleproto.GossipVote{
		ValidatorIndex:  validatorIndex,
		SignType:        oracleInfo.PubKey.Type(),
		SignedTimestamp: time.Now().Unix(),
		Votes:           votes,
		PublicKey:       oracleInfo.PubKey.Bytes(),
	}

	address := oracleInfo.PubKey.Address().String()
	// need to mutex lock as it will clash with concurrent gossip
	oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
	currentGossipVote, ok := oracleInfo.GossipVoteBuffer.Buffer[address]
	if ok {
		// append existing entry in gossipVoteBuffer
		newGossipVote.Votes = append(currentGossipVote.Votes, newGossipVote.Votes...)
	}

	// signing of vote should append the signature field of gossipVote
	if err := oracleInfo.PrivValidator.SignOracleVote("", newGossipVote); err != nil {
		log.Errorf("error signing oracle votes")
		// unlock here to prevent deadlock
		oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
		return
	}

	log.Infof("THIS IS MY COMET PUB KEY: %v", oracleInfo.PubKey)

	oracleInfo.GossipVoteBuffer.Buffer[address] = newGossipVote
	oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
}

func contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func PruneGossipVoteBuffer(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	go func(oracleInfo *types.OracleInfo) {
		maxGossipVoteAge := oracleInfo.Config.MaxGossipVoteAge
		if maxGossipVoteAge == 0 {
			maxGossipVoteAge = 3
		}
		ticker := time.Tick(1 * time.Second)
		for range ticker {
			lastHeight := consensusState.GetLastHeight()
			lastBlockTime := consensusState.GetState().LastBlockTime
			log.Infof("last height: %v, last time: %v", lastHeight, lastBlockTime)

			if !contains(oracleInfo.BlockTimestamps, lastBlockTime.Unix()) {
				oracleInfo.BlockTimestamps = append(oracleInfo.BlockTimestamps, lastBlockTime.Unix())
			}

			if len(oracleInfo.BlockTimestamps) < 3 {
				continue
			}

			if len(oracleInfo.BlockTimestamps) > 3 {
				oracleInfo.BlockTimestamps = oracleInfo.BlockTimestamps[1:]
			}

			oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
			var verifyWg sync.WaitGroup

			// prune votes that are older than the maxGossipVoteAge (in terms of block height)
			for valAddr, gossipVote := range oracleInfo.GossipVoteBuffer.Buffer {
				verifyWg.Add(1)
				go func(valAddr string, gossipVote *oracleproto.GossipVote) {
					defer verifyWg.Done()

					newVotes := []*oracleproto.Vote{}
					for _, vote := range gossipVote.Votes {
						if vote.Timestamp >= oracleInfo.BlockTimestamps[0] {
							newVotes = append(newVotes, vote)
						} else {
							log.Infof("deleting vote: %v, from val addr buffer: %v", vote, valAddr)
						}
					}
					gossipVote.Votes = newVotes
					oracleInfo.GossipVoteBuffer.Buffer[valAddr] = gossipVote
				}(valAddr, gossipVote)
			}

			verifyWg.Wait()
			oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
		}
	}(oracleInfo)
}

// Run run oracles
func Run(oracleInfo *types.OracleInfo, consensusState *cs.State) {
	log.Info("[oracle] Service started.")
	RunProcessSignVoteQueue(oracleInfo, consensusState)
	PruneGossipVoteBuffer(oracleInfo, consensusState)
	// start to take votes from app
	for {
		res, err := oracleInfo.ProxyApp.PrepareOracleVotes(context.Background(), &abcitypes.RequestPrepareOracleVotes{})
		if err != nil {
			log.Errorf("app not ready: %v, retrying...", err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Infof("VOTE: %v", res.Vote)

		oracleInfo.SignVotesChan <- res.Vote
	}
}
