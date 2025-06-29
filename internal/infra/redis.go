package infra

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

type RedisClient struct {
	redisClient *redis.Client
}

func NewRedisClient() *RedisClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	return &RedisClient{
		redisClient: redisClient,
	}
}

func (r *RedisClient) SAdd(ctx context.Context, key string, value string) (int64, error) {
	return r.redisClient.SAdd(ctx, key, value).Result()
}

func (r *RedisClient) SetNx(ctx context.Context, key string, value, ttl time.Duration) (bool, error) {
	return r.redisClient.SetNX(ctx, key, value, ttl).Result()
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.redisClient.Incr(ctx, key).Result()
}

func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return r.redisClient.Decr(ctx, key).Result()
}

func (r *RedisClient) Del(ctx context.Context, key string) (int64, error) {
	return r.redisClient.Del(ctx, key).Result()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.redisClient.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisClient) HSet(ctx context.Context, key string, values map[string]interface{}) (int64, error) {
	return r.redisClient.HSet(ctx, key, values).Result()
}

func (r *RedisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.redisClient.HExists(ctx, key, field).Result()
}

func (r *RedisClient) FlushAll(ctx context.Context) error {
	return r.redisClient.FlushAll(ctx).Err()
}
