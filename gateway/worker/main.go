package main

import (
	"log"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/hibiken/asynq"
	"github.com/storage-gateway/src/config"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/worker/handler"
)

func main() {
	vips.Startup(nil)
	defer vips.Shutdown()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.GetSafeEnv(config.AsynqRedisUrl)},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeBackupFile, handler.HandleBackupTask)
	mux.HandleFunc(queue.TypeUploadFile, handler.HandleUploadTask)
	mux.HandleFunc(queue.TypeDeleteFile, handler.HandleDeleteTask)
	mux.HandleFunc(queue.TypeGenerateThumb, handler.HandleGenerateThumbTask)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
