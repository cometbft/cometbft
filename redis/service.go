package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// Service is a wrapper around redis.Client
type Service struct {
	Client *Client
}

// NewService constructor for Service
func NewService(db int) Service {
	return Service{
		Client: ConnectClient(db),
	}
}

// Exists checks if a redis key exists
func (s Service) Exists(key string) (bool, error) {
	exists, err := s.Client.Exists(key).Result()
	if err != nil {
		return false, err
	}
	return exists != 0, nil
}

// Get get redis value
func (s Service) Get(key string) (GenericValue, bool, error) {
	value, err := s.Client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return GenericValue{}, false, nil
		}
		return GenericValue{}, false, err
	}
	return StringToGenericValue(value), true, nil
}

// Del clears the value for the given key
func (s Service) Del(key string) error {
	err := s.Client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}

// Set sets redis value
func (s Service) Set(key string, value GenericValue, t time.Duration) error {
	err := s.Client.Set(key, value.String(), t).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s Service) SetNX(key string, value GenericValue, t time.Duration) error {
	setnx := s.Client.SetNX(key, value.String(), t)
	err := setnx.Err()
	if err != nil {
		return err
	}
	res, err := setnx.Result()
	if err != nil {
		return err
	}
	if !res { // already exist
		return fmt.Errorf("SetNX for key: %s exists", key)
	}
	return nil
}
