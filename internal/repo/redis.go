package repo

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
)

type RedisClient struct {
	redisClient *redis.Client
}

func NewRedisClient() *RedisClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	return &RedisClient{
		redisClient: redisClient,
	}
}

func (r *RedisClient) SAdd(ctx context.Context, key string, value string) {
	r.redisClient.SAdd(ctx, key, value)
}
