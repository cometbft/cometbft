package adapters

import (
	"encoding/hex"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	"github.com/holiman/uint256"
)

// EVMValueParser struct for evmValueParser
type EVMValueParser struct {
	redisService *redis.Service
}

func NewEVMValueParser(redisService *redis.Service) *EVMValueParser {
	return &EVMValueParser{
		redisService: redisService,
	}
}

// Id returns evmValueParser Id
func (evmValueParser *EVMValueParser) Id() string {
	return "evm_value_parser"
}

// Validate validate job config
func (evmValueParser *EVMValueParser) Validate(job types.OracleJob) error {
	outputType := job.ConfigValue("output_type")
	switch {
	case outputType.IsEmpty():
		return fmt.Errorf("%s: outputType cannot be blank", job.OutputId)
	}
	return nil
}

func ParseSingleEvmResponse(outputType string, evmResponse []byte) (redis.GenericValue, error) {
	switch outputType {
	case "uint256":
		z := new(big.Int)
		z.SetBytes(evmResponse)
		_, overflow := uint256.FromBig(z)
		if overflow {
			panic("number is more than uint256")
		}
		sdkInt := sdkmath.NewIntFromBigInt(z)
		return types.StringToGenericValue(sdkInt.String()), nil
	default:
		panic("unsupported operation: " + outputType)
	}
}

// Perform performs a network request
func (evmValueParser *EVMValueParser) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	rspHexStr := job.GetInput(result).String()
	rsp, err := hex.DecodeString(rspHexStr)
	if err != nil {
		return result, err
	}
	outputType := job.ConfigValue("output_type").String()

	genericValue, err := ParseSingleEvmResponse(outputType, rsp)
	if err != nil {
		return result, err
	}

	result = job.SetOutput(result, genericValue)
	return result, nil
}
