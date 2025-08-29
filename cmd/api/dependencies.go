package main

import (
	"github.com/go-playground/validator/v10"
	httpHandlers "github.com/grcflEgor/go-anagram-api/internal/controller/http/v1"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/internal/storage"
	"github.com/grcflEgor/go-anagram-api/internal/worker"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/patrickmn/go-cache"
)

type Dependencies struct {
	Config         *Config
	Cache          *cache.Cache
	Validator      *validator.Validate
	TaskStorage    storage.TaskStorage
	AnagramService service.AnagramServiceProvider
	WorkerPool     *worker.Pool
	TaskQueue      chan *domain.Task
	Handlers       *httpHandlers.Handlers
}

func NewDependencies(config *Config) *Dependencies {
	appCache := cache.New(config.Cache.DefaultExpiration, config.Cache.CleanupInterval)

	appValidator := validator.New()

	inMemoryStorage := storage.NewInMemoryStorage()
	cachedTaskRepository := storage.NewCachedTaskRepository(inMemoryStorage, appCache)

	taskQueue := make(chan *domain.Task, config.Task.QueueSize)

	anagramService := service.NewAnagramService(cachedTaskRepository, taskQueue)

	workerPool := worker.NewPool(cachedTaskRepository, taskQueue, logger.AppLogger, config.Processing.Timeout)

	handlers := httpHandlers.NewHandlers(anagramService, appValidator)

	return &Dependencies{
		Config:         config,
		Cache:          appCache,
		Validator:      appValidator,
		TaskStorage:    cachedTaskRepository,
		AnagramService: anagramService,
		WorkerPool:     workerPool,
		TaskQueue:      taskQueue,
		Handlers:       handlers,
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
