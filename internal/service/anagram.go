package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ AnagramServiceProvider = (*AnagramService)(nil)

type AnagramService struct {
	storage   storage.TaskStorage
	taskQueue chan<- *domain.Task
	taskStats *TaskStats
}

func NewAnagramService(storage storage.TaskStorage, taskQueue chan<- *domain.Task, taskStats *TaskStats) *AnagramService {
	return &AnagramService{
		storage:   storage,
		taskQueue: taskQueue,
		taskStats: taskStats,
	}
}

func (as *AnagramService) CreateTask(ctx context.Context, words []string, caseSensitive bool) (string, error) {
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

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(task.TraceContext))

	if err := as.storage.Save(ctx, task); err != nil {
		span.RecordError(err)
		return "", err
	}

	as.taskQueue <- task
	as.taskStats.IncrementTotalTasks()
	return task.ID, nil
}

func (as *AnagramService) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	tr := otel.Tracer("usecase")
	ctx, span := tr.Start(ctx, "GetTaskByID")
	defer span.End()

	task, err := as.storage.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
	}
	return task, err
}

func (as *AnagramService) ClearCache(ctx context.Context) error {
	tr := otel.Tracer("usecase")
	ctx, span := tr.Start(ctx, "ClearCache")
	defer span.End()

	if flusher, ok := as.storage.(interface {Flush(ctx context.Context) error}); ok {
		return flusher.Flush(ctx)
	}
	return nil
}