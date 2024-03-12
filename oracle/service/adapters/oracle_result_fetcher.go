package adapters

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/proto/tendermint/oracle"
	"github.com/cometbft/cometbft/redis"
	"github.com/sirupsen/logrus"
)

const ORACLE_ID = "oracle_id"
const STALE_ALLOWANCE = "stale_allowance"

// OracleResultFetcher struct for float handler
type OracleResultFetcher struct {
	redisService *redis.Service
	restUrl      string
}

func NewOracleResultFetcher(redisService *redis.Service, restUrl string) *OracleResultFetcher {
	return &OracleResultFetcher{
		redisService: redisService,
		restUrl:      restUrl,
	}
}

// Id returns cache fetcher Id
func (oracleResultFetcher *OracleResultFetcher) Id() string {
	return "oracle_result_fetcher"
}

// Validate validate job config
func (oracleResultFetcher *OracleResultFetcher) Validate(job types.OracleJob) error {
	oracleId := job.ConfigValue(ORACLE_ID)
	staleAllowance := job.ConfigValue(STALE_ALLOWANCE)
	if oracleId.IsEmpty() {
		return fmt.Errorf("oracle ID cannot be blank")
	}
	if staleAllowance.IsEmpty() {
		return fmt.Errorf("stale allowance cannot be blank")
	}
	return nil
}

// Perform handles cache fetcher operations
func (oracleResultFetcher *OracleResultFetcher) Perform(job types.OracleJob, result types.AdapterResult, _ types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	oracleID := job.ConfigValue(ORACLE_ID).String()
	staleAllowance := job.ConfigValue(STALE_ALLOWANCE).String()

	price, cacheErr := getOracleResultFromCache(oracleID, staleAllowance, *oracleResultFetcher.redisService)

	if cacheErr != nil {
		logrus.Error(cacheErr)
		var apiErr error
		price, apiErr = getOracleResultFromAPI(oracleID, oracleResultFetcher.restUrl)
		if apiErr != nil {
			return result, apiErr
		}
	}

	job.SetOutput(result, types.StringToGenericValue(price))

	return result, nil
}

// GetOracleResultKey returns the redis key for a given oracle
func GetOracleResultKey(oracleId string) string {
	return fmt.Sprintf("oracle-result:%s", oracleId)
}

func getOracleResultFromCache(oracleId string, staleAllowance string, redisService redis.Service) (string, error) {
	key := GetOracleResultKey(oracleId)
	outputGeneric, ok, err := redisService.Get(key)
	if err != nil || !ok {
		return "", err
	}

	outputString := outputGeneric.String()
	var oracleCache types.OracleCache
	unmarshalErr := json.Unmarshal([]byte(outputString), &oracleCache)
	if unmarshalErr != nil {
		return "", unmarshalErr
	}

	elapsedTime := time.Since(oracleCache.Timestamp.Time)
	timeout, err := strconv.ParseUint(staleAllowance, 10, 64)
	if err != nil {
		return "", err
	}

	if elapsedTime.Seconds() > float64(timeout) {
		return "", fmt.Errorf("oracle: %s stale allowance exceeded", oracleId)
	}

	return oracleCache.Price, nil
}

func getOracleResultFromAPI(oracleID, restUrl string) (string, error) {
	if restUrl == "" {
		restUrl = "https://test-api.carbon.network"
	}
	restUrl = restUrl + "/carbon/oracle/v1/results/" + oracleID
	response := HTTPRequest(restUrl, 10)

	if len(response) == 0 {
		return "", fmt.Errorf("empty response from %s", restUrl)
	}

	type Response struct {
		Results []oracle.Result `json:"results"`
	}

	var parsedResponse Response

	if err := json.Unmarshal(response, &parsedResponse); err != nil {
		return "", err
	}

	grpcResult := []oracle.Result{parsedResponse.Results[len(parsedResponse.Results)-1]}

	return grpcResult[0].Data, nil
}
