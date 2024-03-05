package adapters

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
)

// DecimalHandler struct for decimal handler
type DecimalHandler struct {
	redisService *redis.Service
}

func NewDecimalHandler(redisService *redis.Service) *DecimalHandler {
	return &DecimalHandler{
		redisService: redisService,
	}
}

// Id returns float handler Id
func (handler *DecimalHandler) Id() string {
	return "decimal_handler"
}

// Validate validate job config
func (handler *DecimalHandler) Validate(job types.OracleJob) error {
	op := job.ConfigValue("operation").String()
	switch op {
	case "decrease", "increase":
		// check the exponent is a valid uint
		job.ConfigValue("exponent").Uint64()
	default:
		return fmt.Errorf("unsupported operation '%s'", op)
	}
	return nil
}

// ShiftDecimalLeft - Shift decimal point to the left by exponent
func ShiftDecimalLeft(value sdkmath.LegacyDec, exponent uint64) string {
	multiplier := sdkmath.NewIntWithDecimal(1, int(exponent))
	return value.QuoInt(multiplier).String()
}

// ShiftDecimalRight - Shift decimal point to the right by exponent
func ShiftDecimalRight(value sdkmath.LegacyDec, exponent uint64) string {
	multiplier := sdkmath.NewIntWithDecimal(1, int(exponent))
	return value.MulInt(multiplier).String()
}

// Perform handles float operations
func (handler *DecimalHandler) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	input := job.GetInput(result)
	if input.IsEmpty() {
		return result, fmt.Errorf("%s: input cannot be empty", job.InputId)
	}
	inputDec, err := sdkmath.LegacyNewDecFromStr(input.String())
	if err != nil {
		return result, err
	}

	var output redis.GenericValue
	op := job.ConfigValue("operation").String()
	exponent := job.ConfigValue("exponent").Uint64()

	switch op {
	case "decrease":
		output = types.StringToGenericValue(ShiftDecimalLeft(inputDec, exponent))
	case "increase":
		output = types.StringToGenericValue(ShiftDecimalRight(inputDec, exponent))
	default:
		panic("unsupported operation: " + op)
	}

	job.SetOutput(result, output)

	return result, nil
}
