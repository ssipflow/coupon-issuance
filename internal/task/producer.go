package task

import (
	"github.com/hibiken/asynq"
	"os"
)

func JobProduce(key string, payload []byte) error {
	task := asynq.NewTask(key, payload)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	defer client.Close()

	if _, err := client.Enqueue(task); err != nil {
		return err
	}
	return nil
}
