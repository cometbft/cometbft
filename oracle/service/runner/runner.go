package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cometbft/cometbft/oracle/service/adapters"
	"github.com/cometbft/cometbft/oracle/service/parser"
	"github.com/cometbft/cometbft/oracle/service/types"

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
	oraclesURL := oracleInfo.Config.RestApiAddress
	if oraclesURL == "" {
		oraclesURL = "https://test-api.carbon.network"
	}
	oraclesURL += "/carbon/oracle/v1/oracles"

	response := adapters.HTTPRequest(oraclesURL, 10)

	if len(response) == 0 {
		return nil, fmt.Errorf("empty response from %s", oraclesURL)
	}

	type Response struct {
		Oracles []oracleproto.Oracle `json:"oracles"`
	}

	var parsedResponse Response

	if err := json.Unmarshal(response, &parsedResponse); err != nil {
		return nil, err
	}

	oraclesData := parsedResponse

	for _, oracle := range oraclesData.Oracles {
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

		resoUint64, err := strconv.ParseUint(oracle.Resolution, 10, 64)
		if err != nil {
			log.Warnf("[oracle: %v,] unable to parse reso to uint64: %v", oracle.Id, err)
			continue
		}

		oracles = append(oracles, types.Oracle{
			Id:         oracle.Id,
			Resolution: resoUint64,
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
		Config:    oracleInfo.CustomNodeConfig,
	}

	for _, job := range jobs {
		adapter, ok := oracleInfo.AdapterMap[job.Adapter]
		if !ok {
			// panic("adapter should exist: " + job.Adapter)
			return fmt.Errorf("invalid adapter: %s, skipping oracle :%s", job.Adapter, oracle.Id)
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

			// prune gossip vote that have signed timestamps older than 60 secs
			for valAddr, gossipVote := range oracleInfo.GossipVoteBuffer.Buffer {
				if gossipVote.SignedTimestamp < currTime-uint64(interval) {
					delete(oracleInfo.GossipVoteBuffer.Buffer, valAddr)
				}
			}
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
	waitForGrpc(oracleInfo.Config.GrpcAddress)
	waitForRestAPI(oracleInfo.Config.RestApiAddress)
	count := 0
	RunProcessSignVoteQueue(oracleInfo)
	PruneUnsignedVoteBuffer(oracleInfo)
	PruneGossipVoteBuffer(oracleInfo)
	for {
		if count == 0 { // on init, and every minute
			oracles, err := SyncOracles(oracleInfo)
			oracleInfo.GossipVoteBuffer.UpdateMtx.RLock()
			for address := range oracleInfo.GossipVoteBuffer.Buffer {
				log.Infof("THIS IS MY VALIDATOR ADDRESS: %s \n\n\n", address)
			}
			oracleInfo.GossipVoteBuffer.UpdateMtx.RUnlock()
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

		res := adapters.HTTPRequest(address, 10)
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
