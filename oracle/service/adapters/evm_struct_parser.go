package adapters

import (
	"encoding/hex"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"google.golang.org/grpc"
)

// EVMStructParser struct for evmStructParser
type EVMStructParser struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewEVMStructParser(grpcClient *grpc.ClientConn, redisService *redis.Service) *EVMStructParser {
	return &EVMStructParser{
		grpcClient:   grpcClient,
		redisService: redisService,
	}
}

// Id returns evmStructParser Id
func (evmStructParser *EVMStructParser) Id() string {
	return "evm_struct_parser"
}

// Validate validate job config
func (evmStructParser *EVMStructParser) Validate(job types.OracleJob) error {
	outputStruct := job.ConfigValue("output_struct")
	outputType := job.ConfigValue("output_type")
	outputIndex := job.ConfigValue("output_index")
	switch {
	case outputStruct.IsEmpty():
		return fmt.Errorf("%s: output_struct cannot be blank", job.OutputId)
	case outputIndex.IsEmpty():
		return fmt.Errorf("%s: output_index cannot be blank", job.OutputId)
	case outputType.IsEmpty():
		return fmt.Errorf("%s: output_type cannot be blank", job.OutputId)
	}
	return nil
}

func ParseEvmStructResponse(outputType string, outputStruct []string, outputIndex int64, evmResponse []byte) (redis.GenericValue, error) {
	args := abi.Arguments{}
	for _, argType := range outputStruct {
		abiType, err := abi.NewType(argType, "", nil)
		if err != nil {
			panic(err)
		}
		args = append(args, abi.Argument{Type: abiType})
	}
	responseInterface, err := args.Unpack(evmResponse)
	if err != nil {
		return types.StringToGenericValue(""), err
	}
	resultInterface := responseInterface[outputIndex]
	switch outputType {
	case "uint256":
		z := new(big.Int)
		z.SetString(fmt.Sprint(resultInterface), 10)
		sdkInt := sdkmath.NewIntFromBigInt(z)
		return types.StringToGenericValue(sdkInt.String()), nil
	default:
		return types.StringToGenericValue(""), fmt.Errorf("unsupported operation: %s", outputType)
	}
}

// Perform performs a network request
func (evmStructParser *EVMStructParser) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	rspHexStr := job.GetInput(result).String()
	rsp, err := hex.DecodeString(rspHexStr)
	if err != nil {
		return result, err
	}
	outputType := job.ConfigValue("output_type").String()
	outputStruct := job.ConfigValue("output_struct").StringArray()
	outputIndex := job.ConfigValue("output_index").Int64()

	genericValue, err := ParseEvmStructResponse(outputType, outputStruct, outputIndex, rsp)
	if err != nil {
		return result, err
	}

	result = job.SetOutput(result, genericValue)
	return result, nil
}
