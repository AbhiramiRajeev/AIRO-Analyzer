package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/AbhiramiRajeev/AIRO-Analyzer/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
	cfg config.Config // context is used to manage the lifetime of requests to Redis, allowing you to cancel them if needed.
}

func NewRedisClient(add string, password string) (RedisClient, error) {
	client := redis.NewClient(&redis.Options{ // redis.Options is a struct defined in the go-redis library.It holds all the configuration fields you need to connect to a Redis server.
		Addr:     add,
		Password: password,
		DB:       0,
	})
	return RedisClient{
		client: client,
		ctx:    context.Background(), // context.Background() returns a non-nil, empty context. It is used as the top-level context for your application.
	}, nil
}

func (r *RedisClient) AddData(username string, timestamp float64) error {
	// Use the ZAdd command to add a member with a score to a sorted set
	// The score is the timestamp, and the member is the username
	key := fmt.Sprintf("failure-%s", username)
	err := r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  timestamp,
		Member: timestamp,
	})
	if err != nil {
		return fmt.Errorf("error adding data to Redis: %v", err)
	}
	return nil

}

func (r *RedisClient) RemOldFailues(username string, timstamp float64) error {
	key := fmt.Sprintf("failure-%s", username)
	timeVal := float64(time.Now().Unix()) - float64(r.cfg.WindowSize)
	err := r.client.ZRemRangeByScore(r.ctx, key, "-inf", fmt.Sprintf("%f", timeVal)) // ZRemRangeByScore removes all members in the sorted set stored at key with a score between min and max. -inf and +inf are special values that represent negative and positive infinity, respectively. This means it will remove all members with a score less than or equal to the specified timestamp.
	if err != nil {
		return fmt.Errorf("error removing old failures from Redis: %v", err)
	}
	return nil
}
func (r *RedisClient) GetFailedCount(username string) (int, error) {
	key := fmt.Sprintf("failure-%s", username)
	count, err := r.client.ZCount(r.ctx, key, "-inf", "+inf").Result() // ZCount returns the number of elements in the sorted set stored at key with a score between min and max.
	if err != nil {
		return 0, fmt.Errorf("error getting failed count from Redis: %v", err)
	}
	return int(count), nil
}

func (r *RedisClient) AddSuspiciousIp(ip string) error {
	err := r.client.SAdd(r.ctx, "suspicious_ips", ip).Err() // SAdd adds the specified members to the set stored at key. If key does not exist, a new set is created before adding the specified members.
	if err != nil {
		return fmt.Errorf("error adding suspicious IP to Redis: %v", err)
	}
	return nil
}

func (r *RedisClient) IsSuspiciousIp(ip string) (bool, error) {
	isMember, err := r.client.SIsMember(r.ctx, "suspicious_ips", ip).Result() // SIsMember returns true if the member is part of the set, false otherwise.
	if err != nil {
		return false, fmt.Errorf("error checking if IP is suspicious: %v", err)
	}
	return isMember, nil
}

func (r *RedisClient) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("error closing Redis client: %v", err)
	}
	return nil
}
