package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/repository"
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
	task := &domain.Task{
		ID: uuid.New().String(),
		Status: domain.StatusProcessing,
		Words: words,
		CreatedAt: time.Now(),
	}

	if err := uc.repo.Save(ctx, task); err != nil {
		return "", err
	}

	uc.taskQueue <- task

	return task.ID, nil
}

func (uc *AnagramUseCase) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	return uc.repo.GetByID(ctx, id)
}