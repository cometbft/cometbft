package adapters

import (
	"fmt"
	neturl "net/url"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
)

// FetcherMultiple struct for fetcherMultiple
type FetcherMultiple struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewFetcherMultiple(grpcClient *grpc.ClientConn, redisService *redis.Service) *FetcherMultiple {
	return &FetcherMultiple{
		grpcClient:   grpcClient,
		redisService: redisService,
	}
}

// Id returns fetcherMultiple Id
func (fetcherMultiple *FetcherMultiple) Id() string {
	return "fetcher_multiple"
}

// Validate validate job config
func (fetcherMultiple *FetcherMultiple) Validate(job types.OracleJob) error {
	node_key := job.ConfigValue("custom_node_url_key")
	timeout := job.ConfigValue("timeout")
	paths := job.ConfigValue("paths").StringArray()
	switch {
	case node_key.IsEmpty():
		return fmt.Errorf("custom_node_url_key cannot be blank")
	case timeout.Uint64() == 0:
		return fmt.Errorf("invalid timeout")
	case len(paths) == 0:
		return fmt.Errorf("no paths")
	}
	return nil
}

func GetUrl(job types.OracleJob, config types.Config) (string, error) {
	nodeHost := job.ConfigValue("default_node_host").String()
	nodePath := job.ConfigValue("default_node_path").String()
	nodeKey := job.ConfigValue("custom_node_url_key").String()

	customNode, exists := config[nodeKey]

	if exists && customNode.Host != "" && customNode.Path != "" {
		nodeHost = customNode.Host
		nodePath = customNode.Path
		baseUrl, err := neturl.Parse(nodeHost)
		if err != nil {
			log.Warnf("%s: fetcher_multiple: error parsing custom url, using default url instead: %s", nodeKey, err)
		} else {
			url := baseUrl.String() + nodePath
			return url, nil
		}
	}

	baseUrl, err := neturl.Parse(nodeHost)
	if err != nil {
		return "", fmt.Errorf("fetcher_multiple: error parsing default url: %s", err)
	}
	url := baseUrl.String() + nodePath

	return url, nil
}

// Perform performs a network request
func (fetcherMultiple *FetcherMultiple) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	url, err := GetUrl(job, runTimeInput.Config)
	if err != nil {
		return result, err
	}

	timeout := job.ConfigValue("timeout").Uint64()
	reqBody := job.ConfigValue("request_body").String()

	responseStr := GetUrlResponse(url, timeout, reqBody)
	if responseStr == "" {
		return result, fmt.Errorf("empty response from %s", url)
	}

	paths := job.ConfigValue("paths").StringArray()
	outputs := []string{}
	for idx := range paths {
		output := ""
		if paths[idx] != "" {
			value := gjson.Get(responseStr, paths[idx])
			output = value.String()
		}

		if output == "" {
			result = job.SetOutput(result, types.StringToGenericValue(""))
			return result, fmt.Errorf("empty output for %s at %s", url, paths[idx])
		}
		outputs = append(outputs, output)
	}
	if len(paths) == 1 {
		result = job.SetOutput(result, types.StringToGenericValue(outputs[0]))
		return result, nil
	}
	result = job.SetOutputList(result, outputs)
	return result, nil
}
