package repo

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
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

func (r *RedisClient) SetNx(ctx context.Context, key string, value, ttl time.Duration) (bool, error) {
	return r.redisClient.SetNX(ctx, key, value, ttl).Result()
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.redisClient.Incr(ctx, key).Result()
}
