package adapters

import (
	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// GetAdapterMap returns a map of all adapters
func GetAdapterMap(grpcClient *grpc.ClientConn, redisService *redis.Service) map[string]types.Adapter {
	adapterMap := make(map[string]types.Adapter)
	adaptersList := []types.Adapter{
		NewFetcher(grpcClient, redisService), NewFetcherMultiple(grpcClient, redisService),
		NewUnresponsiveHandler(grpcClient, redisService), NewUnchangedHandler(grpcClient, redisService),
		NewMedianFilter(grpcClient, redisService), NewWeightedAverage(grpcClient, redisService),
		NewFloatHandler(grpcClient, redisService), NewDecimalHandler(grpcClient, redisService),
		NewMathFilter(grpcClient, redisService), NewOracleResultFetcher(grpcClient, redisService),
		NewEVMStructParser(grpcClient, redisService), NewEVMFetcher(grpcClient, redisService),
		NewStaticHandler(grpcClient, redisService),
	}
	for _, adapter := range adaptersList {
		_, existing := adapterMap[adapter.Id()]
		if existing {
			log.Errorf("Duplicate ID for %s, ignoring duplicate", adapter.Id())
			continue
		}
		adapterMap[adapter.Id()] = adapter
	}
	return adapterMap
}
