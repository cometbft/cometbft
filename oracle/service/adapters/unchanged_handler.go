package adapters

import (
	"fmt"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// UnchangedHandler struct for unresponsive handler
type UnchangedHandler struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewUnchangedHandler(grpcClient *grpc.ClientConn, redisService *redis.Service) *UnchangedHandler {
	return &UnchangedHandler{
		grpcClient:   grpcClient,
		redisService: redisService,
	}
}

// Id returns unchanged handler Id
func (handler *UnchangedHandler) Id() string {
	return "unchanged_handler"
}

// Validate validate job config
func (handler *UnchangedHandler) Validate(job types.OracleJob) error {
	strategy := job.ConfigValue("strategy").String()
	switch strategy {
	case "nullify":
	default:
		return fmt.Errorf("unsupported strategy '%s'", strategy)
	}
	// check that threshold_duration is a valid uint
	job.ConfigValue("threshold_duration").Uint64()
	return nil
}

// StoreHash struct for store hash
type StoreHash struct {
	Value          string
	LastUpdateTime uint64
}

// Perform handles unchanged responses
func (handler *UnchangedHandler) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, store *types.AdapterStore) (types.AdapterResult, error) {
	input := job.GetInput(result)

	var lastUpdateTime types.GenericValue
	var lastValue types.GenericValue
	if runTimeInput.LastStoreDataExists {
		lastUpdateTime = runTimeInput.GetLastStoreData("last_update_time")
		lastValue = runTimeInput.GetLastStoreData("value")
	}

	thresholdDuration := job.ConfigValue("threshold_duration").Uint64()

	valueWasChanged := input.String() != lastValue.String()
	valueIsStale := lastUpdateTime.Present() && lastUpdateTime.Uint64() < runTimeInput.BeginTime-thresholdDuration

	if runTimeInput.LastStoreDataExists && !valueWasChanged && valueIsStale {
		log.Warnf("[oracle] Unchanged value for %+v, %+v", job.InputId, job.OutputId)
		job.SetOutput(result, types.StringToGenericValue(""))
	} else {
		job.SetOutput(result, input)
	}

	store.ShouldPersist = true

	store.SetData("value", input)
	store.SetData("last_update_time", lastUpdateTime)
	if valueWasChanged {
		store.SetData("last_update_time", types.Uint64ToGenericValue(runTimeInput.BeginTime))
	}

	return result, nil
}
