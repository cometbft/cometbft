package adapters

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
)

// Fetcher struct for fetcher
type Fetcher struct {
	grpcClient   *grpc.ClientConn
	redisService *redis.Service
}

func NewFetcher(grpcClient *grpc.ClientConn, redisService *redis.Service) *Fetcher {
	return &Fetcher{
		grpcClient:   grpcClient,
		redisService: redisService,
	}
}

// Id returns fetcher Id
func (fetcher *Fetcher) Id() string {
	return "fetcher"
}

// Validate validate job config
func (fetcher *Fetcher) Validate(job types.OracleJob) error {
	url := job.ConfigValue("url")
	timeout := job.ConfigValue("timeout")
	path := job.ConfigValue("path")
	switch {
	case url.IsEmpty():
		return fmt.Errorf("url cannot be blank")
	case timeout.Uint64() == 0:
		return fmt.Errorf("invalid timeout")
	case path.IsEmpty():
		return fmt.Errorf("path cannot be blank")
	}
	return nil
}

// Perform performs a network request
func (fetcher *Fetcher) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	url := job.ConfigValue("url").String()
	timeout := job.ConfigValue("timeout").Uint64()
	reqBody := job.ConfigValue("request_body").String()

	responseStr := getUrlResponse(url, timeout, reqBody)
	if responseStr == "" {
		return result, fmt.Errorf("empty response for %s", url)
	}

	output := ""
	path := job.ConfigValue("path").String()
	if path != "" {
		value := gjson.Get(responseStr, job.ConfigValue("path").String())
		output = value.String()
	}

	if output == "" {
		log.Warnln("empty output for " + url)
	}

	result = job.SetOutput(result, types.StringToGenericValue(output))
	return result, nil
}

// getUrlResponse attempts to get a url response from redis cache
// if redis cache is empty, it will make a http call and populate redis with the url as key
func getUrlResponse(url string, timeout uint64, reqBody string) string {
	redService := redis.NewService(0)
	defer redService.Client.Close()
	pool := goredis.NewPool(redService.Client)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	lockKey := GetUrlFetcherLockKey(url)
	locker := rs.NewMutex(lockKey, redsync.WithTries(50), redsync.WithExpiry(time.Second*5), redsync.WithRetryDelay(time.Millisecond*100))

	if err := locker.Lock(); err != nil {
		return ""
	}

	urlResponseKey := GetUrlFetcherResponseKey(url)

	outputGeneric, ok, err := redService.Get(urlResponseKey)
	if err == nil && ok {
		if ok, err := locker.Unlock(); !ok || err != nil {
			panic(fmt.Sprintf("getUrlResponse: unlock failed: %s", err.Error()))
		}
		return outputGeneric.String()
	}

	httpClient := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	var response *http.Response

	//nolint: bodyclose // ignore lint error for response.Body.Close() being outside of if else block
	// if reqBody is present, perform post request instead
	if reqBody != "" {
		response, err = httpClient.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	} else {
		response, err = httpClient.Get(url)
	}

	if err != nil {
		return ""
	}

	defer response.Body.Close()

	body, readErr := ioutil.ReadAll(response.Body)

	if readErr != nil {
		return ""
	}

	output := string(body)

	if err := redService.SetNX(urlResponseKey, types.StringToGenericValue(output), time.Second*5); err != nil {
		log.Warnf("getUrlResponse: failed to set response cache on redis for oracle: %s, %e", urlResponseKey, err)
		return output
	}

	if ok, err := locker.Unlock(); !ok || err != nil {
		log.Warnf("getUrlResponse: failed to unlock response cache mutex for oracle: %v, %s, %e", ok, urlResponseKey, err)
		return output
	}

	return output
}

// GetUrlFetcherLockKey returns the lock key for a given url
func GetUrlFetcherLockKey(url string) string {
	return fmt.Sprintf("oracle:url-lock:%s", url)
}

// GetUrlFetcherResponseKey returns the response key for a given url
func GetUrlFetcherResponseKey(url string) string {
	return fmt.Sprintf("oracle:url-response:%s", url)
}
