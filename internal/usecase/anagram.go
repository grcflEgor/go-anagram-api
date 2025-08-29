package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ AnagramUseCaseProvider = (*AnagramUseCase)(nil)

type AnagramUseCase struct {
	repo repository.TaskRepository
	taskQueue chan<- *domain.Task
}

func NewAnagramUseCase(repo repository.TaskRepository, taskQueue chan<- *domain.Task) *AnagramUseCase {
	return &AnagramUseCase{
		repo: repo,
		taskQueue: taskQueue,
	}
}

func (uc *AnagramUseCase) CreateTask(ctx context.Context, words []string) (string, error) {
	tr := otel.Tracer("usecase")
	ctx, span := tr.Start(ctx, "CreateTask")
	defer span.End()

	task := &domain.Task{
		ID:        uuid.New().String(),
		Status:    domain.StatusProcessing,
		Words:     words,
		CreatedAt: time.Now(),
		TraceContext: make(map[string]string),
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(task.TraceContext))

	if err := uc.repo.Save(ctx, task); err != nil {
		span.RecordError(err)
		return "", err
	}

	uc.taskQueue <- task
	return task.ID, nil
}

func (uc *AnagramUseCase) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	tr := otel.Tracer("usecase")
	ctx, span := tr.Start(ctx, "GetTaskByID")
	defer span.End()

	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
	}
	return task, err
}