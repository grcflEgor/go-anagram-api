package repository

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
)

type TaskRepository interface {
	Save(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id string) (*domain.Task, error)
}