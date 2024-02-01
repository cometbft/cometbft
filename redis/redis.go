package redis

import (
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// Client is a reexport of the redis client
type Client = redis.Client

func GetEnvOr(key string, defaultValue string) string {
	value, found := os.LookupEnv(key)
	if !found {
		return defaultValue
	}
	return value
}

func GetSettings(db int) *redis.Options {
	dbString := GetEnvOr("REDIS_URL", "")
	if dbString != "" {
		url, err := redis.ParseURL(dbString)
		if err != nil {
			panic(err)
		}
		url.DB = 0
		return url
	}

	host := GetEnvOr("REDIS_HOST", "localhost")
	port := GetEnvOr("REDIS_PORT", "6379")
	password := GetEnvOr("REDIS_PASSWORD", "")
	log.Debugf("REDIS_HOST=%v REDIS_HOST=%v REDIS_PASSWORD=%v", host, port, password)

	return &redis.Options{
		Addr:     host + ":" + port,
		Password: password, // no password set
		DB:       db,       // use default DB
	}
}

// ConnectClient returns a redis client
func ConnectClient(db int) *redis.Client {
	client := redis.NewClient(GetSettings(db))

	// Polling for redis
	for {
		err := client.Ping().Err()
		if err == nil {
			break
		}
		if !strings.HasSuffix(err.Error(), ": connect: connection refused") {
			panic("Cannot connect to redis: " + err.Error())
		}
		log.Warn("Redis: Unable to connect, retrying in 5 seconds...")
		time.Sleep(time.Second)
	}

	return client
}
