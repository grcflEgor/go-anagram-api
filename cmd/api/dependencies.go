package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/grcflEgor/go-anagram-api/internal/config"
	httpHandlers "github.com/grcflEgor/go-anagram-api/internal/controller/http/v1"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/internal/storage"
	"github.com/grcflEgor/go-anagram-api/internal/worker"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/patrickmn/go-cache"
)

type Dependencies struct {
	Config         *config.Config
	Cache          *cache.Cache
	Validator      *validator.Validate
	TaskStorage    storage.TaskStorage
	AnagramService service.AnagramServiceProvider
	WorkerPool     *worker.Pool
	TaskQueue      chan *domain.Task
	Handlers       *httpHandlers.Handlers
	TaskStats      *service.TaskStats
}

func NewDependencies(config *config.Config) *Dependencies {
	appCache := cache.New(config.Cache.DefaultExpiration, config.Cache.CleanupInterval)

	appValidator := validator.New()

	inMemoryStorage := storage.NewInMemoryStorage()
	cachedTaskStorage := storage.NewCachedTaskStorage(inMemoryStorage, appCache)

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
