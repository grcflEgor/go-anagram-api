package usecase

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
)

type AnagramUseCaseProvider interface {
	CreateTask(ctx context.Context, words []string) (string, error)
	GetTaskByID(ctx context.Context, id string) (*domain.Task, error)
}