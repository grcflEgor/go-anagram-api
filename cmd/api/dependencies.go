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
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	Config         *config.Config
	Validator      *validator.Validate
	TaskStorage    repositories.TaskStorage
	AnagramService service.AnagramServiceProvider
	WorkerPool     *worker.Pool
	TaskQueue      chan *domain.Task
	Handlers       *httpHandlers.Handlers
	TaskStats      *service.TaskStats
}

func NewDependencies(config *config.Config) *Dependencies {
	appValidator := validator.New()

	pool := mustCreatePgxPool(config)

	redisClient := mustCreateRedisClient(config)

	postgresPool := storage.NewPostgresTaskRepo(pool)

	redisCachedRepo := storage.NewRedisCachedStorage(redisClient, config.Cache.DefaultExpiration)

	//cachedTaskStorage := storage.NewCachedTaskStorage(postgresPool, redisCacheRepo)

	taskQueue := make(chan *domain.Task, config.Task.QueueSize)

	taskStats := service.NewTaskStats()

	anagramService := service.NewAnagramService(postgresPool, redisCachedRepo, taskQueue, taskStats, config.Upload.BatchSize)

	workerPool := worker.NewPool(postgresPool, redisCachedRepo, taskQueue, logger.AppLogger, config.Processing.Timeout, taskStats, config.Upload.BatchSize)

	handlers := httpHandlers.NewHandlers(anagramService, appValidator, config, taskStats)

	return &Dependencies{
		Config:         config,
		Validator:      appValidator,
		TaskStorage:    postgresPool,
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

func mustCreateRedisClient(cfg *config.Config) *redis.Client {
    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        log.Fatal("FATAL: REDIS_URL environment variable is not set")
    }

    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        log.Fatalf("FATAL: Unable to parse REDIS_URL: %v\n", err)
    }

    client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Minute)
	defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        log.Fatalf("FATAL: Failed to connect to Redis: %v\n", err)
    }

    logger.AppLogger.Info("Redis client created and connected")
    return client
}