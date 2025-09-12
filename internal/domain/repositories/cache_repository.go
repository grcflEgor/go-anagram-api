package repositories

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
)


type CacheTaskStorage interface {
	Save(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context) error
}