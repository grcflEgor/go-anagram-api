package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpHandlers "github.com/grcflEgor/go-anagram-api/internal/delivery/http"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/repository"
	"github.com/grcflEgor/go-anagram-api/internal/usecase"
	"github.com/grcflEgor/go-anagram-api/internal/worker"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

const (
	port = ":8080"
	taskQueueSize = 100
	numWorkers = 4
)

func main() {
	logger.InitLogger()
	defer func() { _ = logger.AppLogger.Sync() }()

	appCache := cache.New(5*time.Minute, 10*time.Minute)

	inMemoryRepo := repository.NewInMemoryStorage()
	cachedRepo := repository.NewCachedTaskRepository(inMemoryRepo, appCache)

	taskQueue := make(chan *domain.Task, taskQueueSize)
	anagramUseCase := usecase.NewAnagramUseCase(cachedRepo, taskQueue)

	workerPool := worker.NewPool(cachedRepo, taskQueue, logger.AppLogger)
	workerPool.Run(numWorkers)

	handlers := httpHandlers.NewHandlers(anagramUseCase)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(httpHandlers.LoggerMiddleware)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.HealthCheck)
		r.Post("/anagrams/group", handlers.GroupAnagrams)
		r.Get("/anagrams/groups/{id}", handlers.GetResult)
	})

	srv := &http.Server{
		Addr: port,
		Handler: r,
	}

	go func() {
		logger.AppLogger.Info("Starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.AppLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.AppLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.AppLogger.Fatal("server forced to shutdown", zap.Error(err))
	}

	close(taskQueue)
	logger.AppLogger.Info("task queue closed, waiting for workers finish")
	logger.AppLogger.Info("server exiting")
}
	