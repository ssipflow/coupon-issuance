package task

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/ssipflow/coupon-issuance/internal/infra"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"log"
	"os"
)

type AsynqWorker struct {
	couponRepository *repo.CouponRepository
	redisClient      *infra.RedisClient
}

func NewAsynqWorker(repository *repo.CouponRepository, redisClient *infra.RedisClient) *AsynqWorker {
	return &AsynqWorker{
		couponRepository: repository,
		redisClient:      redisClient,
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
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[ASYNQ] TASK: %s, PAYLOAD: %s, ERROR: %v", task.Type(), string(task.Payload()), err)
			}),
		},
	)

	consumer := NewConsumer(a.redisClient, a.couponRepository)

	mux := asynq.NewServeMux()
	mux.HandleFunc("task:coupon:issue", consumer.IssueCouponProcessor())

	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq worker failed: %v", err)
	}
}
