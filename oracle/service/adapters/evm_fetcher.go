package adapters

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

// EVMFetcher struct for evmFetcher
type EVMFetcher struct {
	redisService *redis.Service
}

func NewEVMFetcher(redisService *redis.Service) *EVMFetcher {
	return &EVMFetcher{
		redisService: redisService,
	}
}

// Id returns evmFetcher Id
func (evmFetcher *EVMFetcher) Id() string {
	return "evm_fetcher"
}

// Validate validate job config
func (evmFetcher *EVMFetcher) Validate(job types.OracleJob) error {
	address := job.ConfigValue("address")
	calldata := job.ConfigValue("calldata")
	nodeRpc := job.ConfigValue("default_node_rpc")
	nodeKey := job.ConfigValue("custom_node_url_key")
	switch {
	case address.IsEmpty():
		return fmt.Errorf("address cannot be blank")
	case calldata.IsEmpty():
		return fmt.Errorf("calldata cannot be blank")
	case nodeRpc.IsEmpty():
		return fmt.Errorf("default_node_rpc cannot be blank")
	case nodeKey.IsEmpty():
		return fmt.Errorf("custom_node_url_key cannot be blank")
	}
	return nil
}

// GetRpcUrl attempts to get the rpc url from config file, if not it populates the config file with the default rpc url
func GetRpcUrl(job types.OracleJob, config types.Config) (string, error) {
	rpcUrl := job.ConfigValue("default_node_rpc").String()
	nodeKey := job.ConfigValue("custom_node_url_key").String()

	customNode, exists := config[nodeKey]

	if exists && customNode.Host != "" {
		rpcUrl = customNode.Host
		url, err := url.Parse(rpcUrl)
		if err != nil {
			log.Warnf("%s: evm_fetcher: error parsing custom rpc url, using default url instead: %s", nodeKey, err)
		} else {
			return url.String(), nil
		}
	}

	url, err := url.Parse(rpcUrl)
	if err != nil {
		return "", fmt.Errorf("evm_fetcher: error parsing default rpc url: %s", err)
	}

	return url.String(), nil
}

// Perform performs a network request
func (evmFetcher *EVMFetcher) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	rpcUrl, err := GetRpcUrl(job, runTimeInput.Config)
	if err != nil {
		return result, err
	}
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return result, err
	}

	addrStr := job.ConfigValue("address").String()
	addr, err := hexutil.Decode(addrStr)
	if err != nil {
		return result, err
	}
	ethAddr := ethcommon.BytesToAddress(addr)

	calldataStr := job.ConfigValue("calldata").String()
	data, err := hexutil.Decode(calldataStr)
	if err != nil {
		return result, err
	}

	msg := ethereum.CallMsg{
		To:   &ethAddr,
		Data: data,
	}
	rsp, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return result, err
	}

	rspHexStr := hex.EncodeToString(rsp)
	genericValue := types.StringToGenericValue(rspHexStr)
	if err != nil {
		return result, err
	}

	result = job.SetOutput(result, genericValue)
	return result, nil
}
