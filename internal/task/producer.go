package task

import (
	"github.com/hibiken/asynq"
	"os"
)

func ProduceTask(taskKey string, payload []byte, opts ...asynq.Option) error {
	task := asynq.NewTask(taskKey, payload)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	defer client.Close()

	if _, err := client.Enqueue(task, opts...); err != nil {
		return err
	}
	return nil
}
