package adapters

import (
	"fmt"

	"github.com/cometbft/cometbft/redis"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/oracle/service/types"
)

// MathFilter struct for float handler
type MathFilter struct {
	redisService *redis.Service
}

func NewMathFilter(redisService *redis.Service) *MathFilter {
	return &MathFilter{
		redisService: redisService,
	}
}

// Id returns math filter Id
func (handler *MathFilter) Id() string {
	return "math_filter"
}

var SUPPORTED_OPS = []string{"divide", "/", "multiply", "*", "add", "+", "subtract", "-", "min", "max"}

// Validate validate job config
func (handler *MathFilter) Validate(job types.OracleJob) error {
	_, err := getOperations(job)
	if err != nil {
		return err
	}
	return nil
}

// getOperations returns the array of operations to be performed.
// Ops can either be a single string or an array of strings, the single string is to keep backwards compatibility,
// specs should use array of strings even if just for a single operation e.g. ["divide"]
func getOperations(job types.OracleJob) ([]string, error) {
	ops := []string{job.ConfigValue("operation").String()}
	if ops[0] == "" || !contains(SUPPORTED_OPS, ops[0]) {
		ops = job.ConfigValue("operation").StringArray()
		for _, op := range ops {
			if !contains(SUPPORTED_OPS, op) {
				return nil, fmt.Errorf("unsupported operation '%s'", op)
			}
		}
	}
	return ops, nil
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// checks if any of the input values for the math_filter job is empty
func isValidInputs(inputs []redis.GenericValue) bool {
	for _, input := range inputs {
		if input.IsEmpty() {
			return false
		}
	}
	return true
}

// Combine combines two feeds into one
// e.g. swth/usdc, osmo/swth => osmo/usdc
func Combine(vals []types.GenericValue, ops []string) (sdkmath.LegacyDec, error) {
	res, err := sdkmath.LegacyNewDecFromStr(vals[0].String())
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}
	for i := 1; i < len(vals); i++ {
		op := ops[i-1]
		val, err := sdkmath.LegacyNewDecFromStr(vals[i].String())
		if err != nil {
			return sdkmath.LegacyDec{}, err
		}
		switch op {
		case "divide", "/":
			if val.IsZero() {
				return sdkmath.LegacyDec{}, fmt.Errorf("val at index %d is zero, cannot divide", i)
			}
			res = res.Quo(val)
		case "multiply", "*":
			res = res.Mul(val)
		case "add", "+":
			res = res.Add(val)
		case "subtract", "-":
			if res.LT(val) {
				return sdkmath.LegacyDec{}, fmt.Errorf("val at index %d is greater than res, cannot subtract", i)
			}
			res = res.Sub(val)
		case "max":
			res = sdkmath.LegacyMaxDec(res, val)
		case "min":
			res = sdkmath.LegacyMinDec(res, val)
		default:
			panic("unsupported operation " + op)
		}
	}
	return res, nil
}

// Perform handles math filter operations
func (handler *MathFilter) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	ops, err := getOperations(job)
	if err != nil {
		return result, fmt.Errorf("%s: unable to get operations for job: %s", job.OutputId, err)
	}

	vals, err := job.GetInputs(result)
	if err != nil {
		return result, fmt.Errorf("%s: unable to get inputs for job: %s", job.OutputId, err)
	}

	if !isValidInputs(vals) {
		return result, fmt.Errorf("%s: empty inputs detected, skipping job", job.OutputId)
	}

	if len(vals) < 2 {
		return result, fmt.Errorf("number of val input %d for math filter not supported", len(vals))
	}

	if len(ops) != len(vals)-1 {
		return result, fmt.Errorf("number of values != number of operations - 1 for oracle_id")
	}

	outputVal, err := Combine(vals, ops)
	if err != nil {
		return result, fmt.Errorf("error combining values: %s", err.Error())
	}
	output := types.StringToGenericValue(outputVal.String())
	job.SetOutput(result, output)
	return result, nil
}
