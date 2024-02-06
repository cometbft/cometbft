package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cometbft/cometbft/oracle/service/types"
	oracle "github.com/cometbft/cometbft/oracle/types"
	"github.com/cometbft/cometbft/redis"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

const ORACLE_ID = "oracle_id"
const STALE_ALLOWANCE = "stale_allowance"

// OracleResultFetcher struct for float handler
type OracleResultFetcher struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewOracleResultFetcher(grpcClient *grpc.ClientConn, redisService *redis.Service) *OracleResultFetcher {
	return &OracleResultFetcher{
		grpcClient:   grpcClient,
		redisService: redisService,
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
func (oracleResultFetcher *OracleResultFetcher) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, store *types.AdapterStore) (types.AdapterResult, error) {
	oracleId := job.ConfigValue(ORACLE_ID).String()
	staleAllowance := job.ConfigValue(STALE_ALLOWANCE).String()

	price, cacheErr := getOracleResultFromCache(oracleId, staleAllowance, *oracleResultFetcher.redisService)

	if cacheErr != nil {
		// rework to re-perform job as we cant use carbon query client due to circular depedency
		logrus.Error(cacheErr)
		var grpcErr error
		price, grpcErr = getOracleResultFromGrpc(oracleId, oracleResultFetcher.grpcClient)
		if grpcErr != nil {
			return result, grpcErr
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

func getOracleResultFromGrpc(oracleId string, grpcClient *grpc.ClientConn) (string, error) {
	oracleClient := oracle.NewQueryClient(grpcClient)
	request := &oracle.QueryResultsRequest{
		OracleId: oracleId,
	}

	// Call the gRPC method to fetch data from the Oracle
	response, err := oracleClient.Results(context.Background(), request)
	if err != nil {
		return "", err
	}

	if len(response.Results) == 0 {
		return "", fmt.Errorf("oracle: %s RPC result is empty", oracleId)
	}

	grpcResult := []oracle.Result{response.Results[len(response.Results)-1]}

	return grpcResult[0].Data, nil
}
