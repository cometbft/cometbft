package adapters

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	"google.golang.org/grpc"
)

// FloatHandler struct for float handler
type FloatHandler struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewFloatHandler(grpcClient *grpc.ClientConn, redisService *redis.Service) *FloatHandler {
	return &FloatHandler{
		grpcClient:   grpcClient,
		redisService: redisService,
	}
}

// Id returns float handler Id
func (handler *FloatHandler) Id() string {
	return "float_handler"
}

// Validate validate job config
func (handler *FloatHandler) Validate(job types.OracleJob) error {
	op := job.ConfigValue("operation").String()
	switch op {
	case "round":
		// check the precision is a valid uint
		job.ConfigValue("precision").Uint64()
	default:
		return fmt.Errorf("unsupported operation: '%s'", op)
	}
	return nil
}

// Round round to the specified precision
func Round(value math.LegacyDec, precision uint64) string {
	multiplier := math.NewIntWithDecimal(1, int(precision))
	val := math.LegacyNewDecFromInt(value.MulInt(multiplier).RoundInt()).QuoInt(multiplier)
	return val.String()
}

// Perform handles float operations
func (handler *FloatHandler) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	input := job.GetInput(result)
	if input.IsEmpty() {
		return result, fmt.Errorf("%s: input cannot be empty", job.InputId)
	}
	value, err := math.LegacyNewDecFromStr(input.String())
	if err != nil {
		return result, err
	}

	var output redis.GenericValue
	op := job.ConfigValue("operation").String()
	switch op {
	case "round":
		rounded := Round(value, job.ConfigValue("precision").Uint64())
		output = types.StringToGenericValue(rounded)
	default:
		panic(fmt.Sprintf("unsupported operation: %s", op))
	}

	job.SetOutput(result, output)

	return result, nil
}
