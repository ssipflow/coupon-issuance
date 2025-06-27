package task

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log"
	"os"
)

func StartWorker(redisClient *redis.Client, db *gorm.DB) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("coupon:issue", IssueCouponProcessor(db, redisClient))

	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq worker failed: %v", err)
	}
}
