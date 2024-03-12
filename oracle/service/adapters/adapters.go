package adapters

import (
	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	log "github.com/sirupsen/logrus"
)

// GetAdapterMap returns a map of all adapters
func GetAdapterMap(redisService *redis.Service, restUrl string) map[string]types.Adapter {
	adapterMap := make(map[string]types.Adapter)
	adaptersList := []types.Adapter{
		NewFetcher(redisService), NewFetcherMultiple(redisService),
		NewUnresponsiveHandler(redisService), NewUnchangedHandler(redisService),
		NewMedianFilter(redisService), NewWeightedAverage(redisService),
		NewFloatHandler(redisService), NewDecimalHandler(redisService),
		NewMathFilter(redisService), NewOracleResultFetcher(redisService, restUrl),
		NewEVMValueParser(redisService), NewEVMStructParser(redisService), NewEVMFetcher(redisService),
		NewStaticHandler(redisService),
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
