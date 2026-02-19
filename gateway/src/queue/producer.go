package queue

import (
	"github.com/hibiken/asynq"
	"github.com/storage-gateway/src/config"
)

var asynqClient *asynq.Client

func InitQueue() *asynq.Client {
	asynqClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr: config.GetSafeEnv(config.AsynqRedisUrl),
	})
	return asynqClient
}
