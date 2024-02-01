package adapters

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	"google.golang.org/grpc"
)

// StaticHandler struct for decimal handler
type StaticHandler struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewStaticHandler(grpcClient *grpc.ClientConn, redisService *redis.Service) *StaticHandler {
	return &StaticHandler{
		grpcClient:   grpcClient,
		redisService: redisService,
	}
}

// Id returns float handler Id
func (handler *StaticHandler) Id() string {
	return "static_handler"
}

// Validate validate job config
func (handler *StaticHandler) Validate(job types.OracleJob) error {
	valStr := job.ConfigValue("value").String()
	_, err := sdkmath.LegacyNewDecFromStr(valStr)
	if err != nil {
		return fmt.Errorf("value %s cannot be cast to sdk.Dec error: %s", valStr, err.Error())
	}
	return nil
}

// Perform handles float operations
func (handler *StaticHandler) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	valStr := job.ConfigValue("value").String()
	val, err := sdkmath.LegacyNewDecFromStr(valStr)
	if err != nil {
		return result, fmt.Errorf("value %s cannot be cast to sdk.Dec error: %s", valStr, err.Error())
	}
	output := types.StringToGenericValue(val.String())

	job.SetOutput(result, output)
	return result, nil
}
