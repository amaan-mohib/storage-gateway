package main

import (
	"log/slog"

	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/storage-gateway/src/internal/config"
	server "github.com/storage-gateway/src/internal/http"
	"github.com/storage-gateway/src/internal/service"
	"github.com/storage-gateway/src/internal/storage/primary"
)

func main() {
	cfg := config.LoadConfig()

	options := s3.Options{
		Region:       cfg.Region,
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: true,
	}

	s3Client := s3.New(options)

	store := primary.NewClient(s3Client)
	files := service.NewFileService(store)
	handler := server.NewHandler(files)

	r := server.NewRouter(handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.ListenAndServe(":5000", r)

	gracefulShutdown(
		// func() error {
		// 	return rabbit.Service.Channel.Close()
		// },
		func() error {
			os.Exit(0)
			return nil
		},
	)
}

func gracefulShutdown(ops ...func() error) {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	if <-shutdown != nil {
		for _, op := range ops {
			if err := op(); err != nil {
				slog.Error("gracefulShutdown op failed", "error", err)
				panic(err)
			}
		}
	}
}
