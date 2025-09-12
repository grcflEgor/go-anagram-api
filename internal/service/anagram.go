package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/domain/repositories"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

var _ AnagramServiceProvider = (*AnagramService)(nil)

type AnagramService struct {
	storage   repositories.TaskStorage
	cache repositories.CacheTaskStorage
	taskQueue chan<- *domain.Task
	taskStats *TaskStats
	batchSize int
}

func NewAnagramService(storage repositories.TaskStorage, cache repositories.CacheTaskStorage, taskQueue chan<- *domain.Task, taskStats *TaskStats, batchSize int) *AnagramService {
	return &AnagramService{
		storage:   storage,
		cache: cache,
		taskQueue: taskQueue,
		taskStats: taskStats,
		batchSize: batchSize,
	}
}

func (as *AnagramService) CreateTask(ctx context.Context, words []string, caseSensitive bool) (string, error) {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("usecase")
	ctx, span := tr.Start(ctx, "CreateTask")
	defer span.End()

	task := &domain.Task{
		ID:           uuid.New().String(),
		Status:       domain.StatusProcessing,
		Words:        words,
		CaseSensitive: caseSensitive,
		CreatedAt:    time.Now(),
		TraceContext: make(map[string]string),
	}

	if len(words) > as.batchSize {
		tmpFile, err := os.CreateTemp("", "anagram-task-*.txt")
		if err != nil {
			span.RecordError(err)
			return "", err
		}
		defer tmpFile.Close()

		w := bufio.NewWriter(tmpFile)
		for _, word := range words {
			fmt.Fprintln(w, word)
		}
		w.Flush()

		task.FilePath = tmpFile.Name()
	} else {
		task.Words = words
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(task.TraceContext))

	if err := as.storage.Save(ctx, task); err != nil {
		span.RecordError(err)
		return "", err
	}

	if err := as.cache.Save(ctx, task); err != nil {
        l.Error("Failed to save task to cache", zap.Error(err))
        span.RecordError(err)
    }

	as.taskQueue <- task
	as.taskStats.IncrementTotalTasks()
	return task.ID, nil
}

func (as *AnagramService) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("usecase")
	ctx, span := tr.Start(ctx, "GetTaskByID")
	defer span.End()

	task, err := as.cache.GetByID(ctx, id)
	if err != nil {
		l.Error("err to get from redis")
		span.RecordError(err)
	}

	if task != nil {
		as.taskStats.IncrementCacheHits()
	}

	as.taskStats.IncrementCacheMiss()
	task, err = as.storage.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if task != nil {
		if err := as.cache.Save(ctx, task); err != nil {
		span.RecordError(err)
		}
	}
	return task, err
}

func (as *AnagramService) ClearCache(ctx context.Context) error {
	tr := otel.Tracer("usecase")
	ctx, span := tr.Start(ctx, "ClearCache")
	defer span.End()

	if err := as.cache.DeleteAll(ctx); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}