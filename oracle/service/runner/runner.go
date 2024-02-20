package runner

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	// "github.com/Switcheo/carbon/constants"
	"github.com/cometbft/cometbft/oracle/service/adapters"
	"github.com/cometbft/cometbft/oracle/service/parser"
	"github.com/cometbft/cometbft/oracle/service/types"
	oracletypes "github.com/cometbft/cometbft/oracle/types"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
	"github.com/cometbft/cometbft/redis"
)

var (
	OracleOverwriteData string
)

// OracleInfoResult oracle info result
type OracleInfoResult struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

// OracleInfoResponse oracle info response
type OracleInfoResponse struct {
	Height string           `json:"height"`
	Result OracleInfoResult `json:"result"`
}

// LastSubmissionTimeKey key for last submission time
const LastSubmissionTimeKey = "oracle:submitter:last-submission-time"

// LastStoreDataKey returns the key for the given adapter and job
func LastStoreDataKey(adapter types.Adapter, job types.OracleJob) string {
	return fmt.Sprintf("oracle:adapter-store:%s:%s", adapter.Id(), job.InputId)
}

// GetLastStoreData returns the last stored value for the given adapter and job
func GetLastStoreData(service redis.Service, adapter types.Adapter, job types.OracleJob) (data map[string]types.GenericValue, exists bool, err error) {
	key := LastStoreDataKey(adapter, job)
	value, exists, err := service.Get(key)
	if err != nil {
		return
	}
	data = make(map[string]types.GenericValue)
	if exists {
		err := json.Unmarshal([]byte(value.String()), &data)
		if err != nil {
			panic(err)
		}
	}
	return data, exists, nil
}

// SetLastStoreData sets the last store data for the given adapter and job
func SetLastStoreData(service redis.Service, adapter types.Adapter, job types.OracleJob, store types.AdapterStore) error {
	key := LastStoreDataKey(adapter, job)
	dataBytes, err := json.Marshal(&store.Data)
	if err != nil {
		panic(err)
	}
	err = service.Set(key, types.StringToGenericValue(string(dataBytes)), 0)
	if err != nil {
		return err
	}
	return nil
}

// GetOracleLockKey returns the lock key for a given oracle and time
func GetOracleLockKey(oracle types.Oracle, normalizedTime uint64) string {
	return fmt.Sprintf("oracle:oracle-lock:%s:%d", oracle.Id, normalizedTime)
}

func overwriteData(oracleId string, data string) string {
	if oracleId != "DXBT" { // if we want to overwrite DETH: `&& oracleID != "DETH"`
		return data
	}

	var min, max, interval int64
	switch oracleId {
	case "DXBT":
		min, max = 15000, 10000 // this was how it was before the refactor, maybe intended?
		interval = 20
	case "DETH":
		min, max = 500, 1500
		interval = 5
	}

	// create a price based on current system time
	t := time.Now().Unix()
	minute := t / 60
	seconds := t - (t/60)*60
	// round to the nearest 10th second, e.g. 10, 20, 30...
	roundedSeconds := (seconds / 10) * 10
	isEvenMinute := minute%2 == 0
	// if the minute is exactly an even minute
	if isEvenMinute {
		if roundedSeconds == 0 {
			return strconv.FormatUint(uint64(min), 10)
		}

		price := strconv.FormatUint(uint64(min+roundedSeconds*interval), 10)
		decimalPrice := strconv.FormatUint(uint64(seconds/10), 10)
		decimalPrice += strconv.FormatUint(10-uint64(seconds/10), 10)
		return price + "." + decimalPrice
	}

	if roundedSeconds == 0 {
		return strconv.FormatUint(uint64(max), 10)
	}

	price := strconv.FormatUint(uint64(max-roundedSeconds*interval), 10)
	decimalPrice := strconv.FormatUint(uint64(seconds/10)+4, 10)
	decimalPrice += strconv.FormatUint(10-uint64(seconds/10), 10)
	return price + "." + decimalPrice
}

// SyncOracles sync oracles with active on-chain oracles
func SyncOracles(oracleInfo *types.OracleInfo) (oracles []types.Oracle, err error) {
	// fetch oracle list first
	oracleClient := oracletypes.NewQueryClient(oracleInfo.GrpcClient)
	oracleRes, err := oracleClient.OracleAll(
		context.Background(),
		&oracletypes.QueryAllOracleRequest{
			//Pagination: &sdkquerytypes.PageRequest{}
		},
	)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	oraclesData := oracleRes.Oracles

	for _, oracle := range oraclesData {
		var spec types.OracleSpec
		err = json.Unmarshal([]byte(oracle.Spec), &spec)
		if err != nil {
			log.Errorf("[oracle:%v] invalid oracle spec: %+v", oracle.Id, err)
			continue
		}
		spec, err := parser.ParseSpec(spec)
		if err != nil {
			log.Warnf("[oracle:%v] unable to unroll spec: %v", oracle.Id, err)
			continue
		}
		err = parser.ValidateOracleJobs(oracleInfo, spec.Jobs)
		if err != nil {
			log.Warnf("[oracle: %v,] invalid oracle jobs: %v", oracle.Id, err)
			continue
		}
		oracles = append(oracles, types.Oracle{
			Id:         oracle.Id,
			Resolution: uint64(oracle.Resolution),
			Spec:       spec,
		})
	}
	log.Info("[oracle] Synced oracle specs")
	return oracles, err
}

func SaveOracleResult(price string, oracleId string, redisService redis.Service) {
	if price != "" {
		key := adapters.GetOracleResultKey(oracleId)
		data, err := json.Marshal(types.OracleCache{Price: price, Timestamp: types.JSONTime{Time: time.Now()}})
		if err != nil {
			panic(err)
		}
		jsonString := string(data)
		setErr := redisService.Set(key, types.StringToGenericValue(jsonString), 0)
		if setErr != nil {
			log.Error(err)
		}
	}
}

// RunOracle run oracle submission
func RunOracle(oracleInfo *types.OracleInfo, oracle types.Oracle, currentTime uint64) error {
	red := oracleInfo.Redis
	normalizedTime := (currentTime / oracle.Resolution) * oracle.Resolution
	lastSubmissionTime, exists, err := red.Get(LastSubmissionTimeKey)
	if err != nil {
		return err
	}
	if exists && normalizedTime <= lastSubmissionTime.Uint64() {
		return nil
	}
	lockKey := GetOracleLockKey(oracle, normalizedTime)
	err = red.SetNX(lockKey, types.StringToGenericValue("1"), time.Minute*5)
	//nolint:nilerr //already processed/processing
	if err != nil {
		return nil
	}

	jobs := oracle.Spec.Jobs
	shouldEarlyTerminate := oracle.Spec.ShouldEarlyTerminate
	result := types.NewAdapterResult()

	input := types.AdapterRunTimeInput{
		BeginTime: currentTime,
		Config:    oracleInfo.Config,
	}

	for _, job := range jobs {
		adapter, ok := oracleInfo.AdapterMap[job.Adapter]
		if !ok {
			panic("adapter should exist: " + job.Adapter)
		}
		input.LastStoreData, input.LastStoreDataExists, err = GetLastStoreData(red, adapter, job)
		if err != nil {
			return err
		}
		store := types.NewAdapterStore()
		result, err = adapter.Perform(job, result, input, &store)
		if err != nil {
			log.Error(fmt.Errorf("%s: %s: %s", oracle.Id, adapter.Id(), err.Error()))
			if shouldEarlyTerminate {
				break
			}
		}
		if store.ShouldPersist {
			if err := SetLastStoreData(red, adapter, job, store); err != nil {
				return err
			}
		}
	}

	err = red.Set(LastSubmissionTimeKey, types.Uint64ToGenericValue(normalizedTime), 0)
	if err != nil {
		return err
	}

	resultData := result.GetData(oracle.Spec.OutputId).String()

	SaveOracleResult(resultData, oracle.Id, red)

	if OracleOverwriteData == "true" {
		resultData = overwriteData(oracle.Id, resultData) // if we want to override oracle price
	}

	if resultData == "" {
		return errors.New("skipping submission for " + oracle.Id + " as result is empty")
	}

	vote := oracleproto.Vote{
		Validator: oracleInfo.PubKey.Address().String(),
		OracleId:  oracle.Id,
		Timestamp: normalizedTime,
		Data:      resultData,
	}

	oracleInfo.SignVotesChan <- &vote
	return nil
}

func RunProcessSignVoteQueue(oracleInfo *types.OracleInfo) {
	go func(oracleInfo *types.OracleInfo) {
		for {
			select {
			case <-oracleInfo.StopChannel:
				return
			default:
				interval := oracleInfo.MsgFlushInterval
				if interval == 0 {
					interval = 500 * time.Millisecond
				}
				time.Sleep(interval)
				ProcessSignVoteQueue(oracleInfo)
			}
		}
	}(oracleInfo)
}

func HashGossipVote(gossipVote *oracleproto.GossipVote) uint64 {
	h := sha256.New()
	encoding, err := gossipVote.Marshal()
	if err != nil {
		panic(err)
	}
	h.Write(encoding)
	hashBytes := h.Sum(nil)
	hashUint64 := binary.BigEndian.Uint64(hashBytes)
	return hashUint64
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
		Validator: oracleInfo.PubKey.Address().String(),
		PublicKey: oracleInfo.PubKey.Bytes(),
		SignType:  oracleInfo.PubKey.Type(),
		Votes:     batchVotes,
	}

	// signing of vote should append the signature field and timestamp field of gossipVote
	if err := oracleInfo.PrivValidator.SignOracleVote("", newGossipVote); err != nil {
		log.Errorf("error signing oracle votes")
	}
	newGossipVote.SignedTimestamp = uint64(time.Now().Unix())

	// replace current gossipVoteBuffer with new one
	address := oracleInfo.PubKey.Address().String()

	// need to mutex lock as it will clash with concurrent gossip
	oracleInfo.GossipVoteBuffer.UpdateMtx.Lock()
	oracleInfo.GossipVoteBuffer.Buffer[address] = newGossipVote
	oracleInfo.GossipVoteBuffer.UpdateMtx.Unlock()
}

func AddVoteToDataBuffer(oracleInfo *types.OracleInfo, vote *oracleproto.Vote) {
	_, ok := oracleInfo.VoteDataBuffer.Buffer[vote.Timestamp]
	if !ok {
		oracleInfo.VoteDataBuffer.Buffer[vote.Timestamp] = make(map[string][]*oracleproto.Vote)
	}

	oracleMap := oracleInfo.VoteDataBuffer.Buffer[vote.Timestamp]
	oracleMap[vote.OracleId] = append(oracleMap[vote.OracleId], vote)
	oracleInfo.VoteDataBuffer.Buffer[vote.Timestamp] = oracleMap
}

func PruneVoteDataBuffer(oracleInfo *types.OracleInfo) {
	go func(oracleInfo *types.OracleInfo) {
		ticker := time.Tick(3 * time.Second)
		for range ticker {
			oracleInfo.VoteDataBuffer.UpdateMtx.Lock()
			// prune everything older than 3 secs
			for timestamp, _ := range oracleInfo.VoteDataBuffer.Buffer {
				currTime := uint64(time.Now().Unix())
				if timestamp <= currTime-uint64(3*time.Second) {
					delete(oracleInfo.VoteDataBuffer.Buffer, timestamp)
				}
			}
			oracleInfo.VoteDataBuffer.UpdateMtx.Unlock()
		}
	}(oracleInfo)
}

func PruneUnsignedVoteBuffer(oracleInfo *types.OracleInfo) {
	go func(oracleInfo *types.OracleInfo) {
		ticker := time.Tick(3 * time.Second)
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
				if timestamp <= currTime-uint64(3*time.Second) {
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

// RunOracles run oracle submissions
func RunOracles(oracleInfo *types.OracleInfo, t uint64) {
	for _, oracle := range oracleInfo.Oracles {
		go func(currOracle types.Oracle) {
			err := RunOracle(oracleInfo, currOracle, t)
			if err != nil {
				log.Warnln(err)
			}
		}(oracle)
	}
}

// Run run oracles
func Run(oracleInfo *types.OracleInfo) {
	log.Info("[oracle] Service started.")
	count := 0
	RunProcessSignVoteQueue(oracleInfo)
	PruneUnsignedVoteBuffer(oracleInfo)
	// PruneVoteDataBuffer(oracleInfo)
	for {
		if count == 0 { // on init, and every minute
			oracles, err := SyncOracles(oracleInfo)
			if err != nil {
				log.Warn(err)
				time.Sleep(time.Second)
				continue
			}
			oracleInfo.Oracles = oracles
		}

		RunOracles(oracleInfo, uint64(time.Now().Unix()))
		time.Sleep(100 * time.Millisecond)

		count++
		if count > 600 { // 600 * 0.1s = 60s = every minute
			count = 0
		}
	}
}
