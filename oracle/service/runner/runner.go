package runner

import (
	"context"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

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
		PublicKey: oracleInfo.PubKey.Bytes(),
		SignType:  oracleInfo.PubKey.Type(),
		Votes:     batchVotes,
	}

	// signing of vote should append the signature field and timestamp field of gossipVote
	if err := oracleInfo.PrivValidator.SignOracleVote("", newGossipVote); err != nil {
		log.Errorf("error signing oracle votes")
	}

	// replace current gossipVoteBuffer with new one
	address := oracleInfo.PubKey.Address().String()

	// need to mutex lock as it will clash with concurrent gossip
	oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
	oracleInfo.GossipVoteBuffer.Buffer[address] = newGossipVote
	oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
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
					log.Infof("DELETING STALE GOSSIP BUFFER FOR VAL: %s", valAddr)
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
			interval = 1 * time.Second
		}
		ticker := time.Tick(interval)
		for range ticker {
			oracleInfo.UnsignedVoteBuffer.UpdateMtx.RLock()
			// prune everything older than 3 secs
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
	waitForGrpc(oracleInfo.Config.GrpcAddress)
	waitForRestAPI(oracleInfo.Config.RestApiAddress)
	RunProcessSignVoteQueue(oracleInfo)
	PruneUnsignedVoteBuffer(oracleInfo)
	PruneGossipVoteBuffer(oracleInfo)
	// start to take votes from app
	for {
		res, err := oracleInfo.ProxyApp.PrepareOracleVotes(context.Background(), &abcitypes.RequestPrepareOracleVotes{})
		if err != nil {
			log.Error(err)
		}

		log.Infof("RESULTS: %v", res.Votes)

		time.Sleep(100 * time.Millisecond)
	}
}

func waitForRestAPI(address string) {
	restMaxRetryCount := 12
	retryCount := 0
	sleepTime := time.Second
	for {
		log.Infof("[oracle] checking if rest endpoint is up %s : %d", address, retryCount)
		if retryCount == restMaxRetryCount {
			panic("failed to connect to grpc:grpcClient after 12 tries")
		}
		time.Sleep(sleepTime)

		res := HTTPRequest(address, 10)
		if len(res) != 0 {
			break
		}

		time.Sleep(time.Duration(retryCount*int(time.Second) + 1))
		retryCount++
		sleepTime *= 2
	}
}

func waitForGrpc(address string) {
	grpcMaxRetryCount := 12
	retryCount := 0
	sleepTime := time.Second
	var client *grpc.ClientConn

	for {
		log.Infof("[oracle] trying to connect to grpc with address %s : %d", address, retryCount)
		if retryCount == grpcMaxRetryCount {
			panic("failed to connect to grpc:grpcClient after 12 tries")
		}
		time.Sleep(sleepTime)

		// reinit otherwise connection will be idle, in idle we can't tell if it's really ready
		var err error
		client, err = grpc.Dial(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(err)
		}
		// give it some time to connect after dailing, but not too long as connection can become idle
		time.Sleep(time.Duration(retryCount*int(time.Second) + 1))

		if client.GetState() == connectivity.Ready {
			break
		}
		client.Close()
		retryCount++
		sleepTime *= 2
	}
}

func HTTPRequest(url string, timeout uint64) []byte {
	httpClient := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	var response *http.Response
	response, err := httpClient.Get(url)

	if err != nil {
		return []byte{}
	}

	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)

	if readErr != nil {
		return []byte{}
	}

	return body
}
