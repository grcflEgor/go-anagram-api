package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/grcflEgor/go-anagram-api/internal/config"
	httpHandlers "github.com/grcflEgor/go-anagram-api/internal/controller/http/v1"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/domain/repositories"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/internal/storage"
	"github.com/grcflEgor/go-anagram-api/internal/worker"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patrickmn/go-cache"
)

type Dependencies struct {
	Config         *config.Config
	Cache          *cache.Cache
	Validator      *validator.Validate
	TaskStorage    repositories.TaskStorage
	AnagramService service.AnagramServiceProvider
	WorkerPool     *worker.Pool
	TaskQueue      chan *domain.Task
	Handlers       *httpHandlers.Handlers
	TaskStats      *service.TaskStats
}

func NewDependencies(config *config.Config) *Dependencies {
	appCache := cache.New(config.Cache.DefaultExpiration, config.Cache.CleanupInterval)

	appValidator := validator.New()

	pool := mustCreatePgxPool(config)

	postgresPool := storage.NewPostgresTaskRepo(pool)
	cachedTaskStorage := storage.NewCachedTaskStorage(postgresPool, appCache)

	taskQueue := make(chan *domain.Task, config.Task.QueueSize)

	taskStats := service.NewTaskStats()

	anagramService := service.NewAnagramService(cachedTaskStorage, taskQueue, taskStats, config.Upload.BatchSize)

	workerPool := worker.NewPool(cachedTaskStorage, taskQueue, logger.AppLogger, config.Processing.Timeout, taskStats, config.Upload.BatchSize)

	handlers := httpHandlers.NewHandlers(anagramService, appValidator, config, taskStats)

	return &Dependencies{
		Config:         config,
		Cache:          appCache,
		Validator:      appValidator,
		TaskStorage:    cachedTaskStorage,
		AnagramService: anagramService,
		WorkerPool:     workerPool,
		TaskQueue:      taskQueue,
		Handlers:       handlers,
		TaskStats:      taskStats,
	}
}

func (d *Dependencies) Start() {
	d.WorkerPool.Run(d.Config.Worker.Count)
	logger.AppLogger.Info("worker pool started")
}

func (d *Dependencies) Stop() {
	d.WorkerPool.Stop()
	logger.AppLogger.Info("worker pool stopped")
}

func mustCreatePgxPool(cfg *config.Config) *pgxpool.Pool {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("FATAL: DATABASE_URL environment variable is not set")
    }

    config, err := pgxpool.ParseConfig(dbURL)
    if err != nil {
        log.Fatalf("FATAL: Unable to parse DATABASE_URL: %v\n", err)
    }

    config.MaxConns = int32(cfg.DBconfig.MaxConns)
    config.MinConns = int32(cfg.DBconfig.MinConns)
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = 30 * time.Minute

    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        log.Fatalf("FATAL: Unable to create connection pool: %v\n", err)
    }
    
    if err := pool.Ping(context.Background()); err != nil {
        log.Fatalf("FATAL: Unable to ping database: %v\n", err)
    }

    logger.AppLogger.Info("Database connection pool created and verified")
    return pool
}