package repo

import (
	"github.com/redis/go-redis/v9"
	"os"
)

func NewRedisClient() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	return redisClient
}
