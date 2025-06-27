package task

import (
	"github.com/hibiken/asynq"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"log"
	"os"
)

type AsynqWorker struct {
	mySqlRepository *repo.MySqlRepository
	redisClient     *repo.RedisClient
}

func NewAsynqWorker(repository *repo.MySqlRepository, redisClient *repo.RedisClient) *AsynqWorker {
	return &AsynqWorker{
		mySqlRepository: repository,
		redisClient:     redisClient,
	}
}

func (a *AsynqWorker) Start() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	consumer := NewConsumer(a.redisClient, a.mySqlRepository)

	mux := asynq.NewServeMux()
	mux.HandleFunc("coupon:issue", consumer.IssueCouponProcessor())

	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq worker failed: %v", err)
	}
}
