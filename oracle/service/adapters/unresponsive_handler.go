package adapters

import (
	"fmt"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	"google.golang.org/grpc"
)

// UnresponsiveHandler struct for unresponsive handler
type UnresponsiveHandler struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewUnresponsiveHandler(grpcClient *grpc.ClientConn, redisService *redis.Service) *UnresponsiveHandler {
	return &UnresponsiveHandler{
		grpcClient:   grpcClient,
		redisService: redisService,
	}
}

// Id returns unresponsive handler Id
func (handler *UnresponsiveHandler) Id() string {
	return "unresponsive_handler"
}

// Validate validate job config
func (handler *UnresponsiveHandler) Validate(job types.OracleJob) error {
	strategy := job.ConfigValue("strategy").String()
	switch strategy {
	case "use_last":
	default:
		return fmt.Errorf("unsupported strategy '%s'", strategy)
	}
	// check that grace_duration is a valid uint
	// TODO: refactor GenericValue to not panic
	job.ConfigValue("grace_duration").Uint64()
	return nil
}

// Perform handles empty responses
func (handler *UnresponsiveHandler) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, store *types.AdapterStore) (types.AdapterResult, error) {
	input := job.GetInput(result)
	output := input

	graceDuration := job.ConfigValue("grace_duration").Uint64()
	var lastResponsiveTime uint64
	if runTimeInput.LastStoreDataExists && runTimeInput.GetLastStoreData("last_responsive_time").Present() {
		lastResponsiveTime = runTimeInput.GetLastStoreData("last_responsive_time").Uint64()
	}

	store.SetData("last_responsive_time", types.Uint64ToGenericValue(lastResponsiveTime))

	withinGraceDuration := lastResponsiveTime > runTimeInput.BeginTime-graceDuration

	if output.IsEmpty() && runTimeInput.LastStoreDataExists && withinGraceDuration {
		output = runTimeInput.GetLastStoreData("value")
	} else {
		store.SetData("last_responsive_time", types.Uint64ToGenericValue(runTimeInput.BeginTime))
	}

	store.ShouldPersist = true
	store.SetData("value", output)

	result = job.SetOutput(result, output)

	return result, nil
}
