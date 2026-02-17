package main

import (
	"log/slog"

	"net/http"
	"os"
	"os/signal"
	"syscall"

	server "github.com/storage-gateway/src/internal/http"
	"github.com/storage-gateway/src/internal/service"
	"github.com/storage-gateway/src/internal/storage/s3_store"
)

func main() {
	store := s3_store.GetPrimaryStore()
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
