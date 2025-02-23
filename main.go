package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mini-ci-runner-go/internal/api"
	"mini-ci-runner-go/internal/logger"
	"mini-ci-runner-go/internal/runner"
	"mini-ci-runner-go/internal/store"
)

func main() {
	// Initialize logger - writes to both console and file
	if err := logger.Init("ci-runner.log"); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	jobStore := store.NewMemoryStore()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := runner.NewWorkerPool(4, jobStore)
	pool.Start(ctx)

	handler := api.NewHandler(jobStore, pool)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler.Routes(),
	}

	go func() {
		logger.Info.Println("Server running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info.Println("Shutting down...")
	cancel()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	server.Shutdown(shutdownCtx)
}
