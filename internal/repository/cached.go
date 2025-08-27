package repository

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var _ TaskRepository = (*CachedTaskRepository)(nil)

type CachedTaskRepository struct {
	next TaskRepository
	c *cache.Cache
}

func NewCachedTaskRepository(next TaskRepository, c *cache.Cache) *CachedTaskRepository {
	return &CachedTaskRepository{
		next: next,
		c: c,
	}
}

func (r *CachedTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	l := logger.FromContext(ctx)

	if task, found := r.c.Get(id); found {
		l.Info("cache HIT for task", zap.String("task_id", id))
		return task.(*domain.Task), nil
	}
	l.Info("cache MISS for task", zap.String("task_id", id))

	task, err := r.next.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.c.Set(id, task, cache.DefaultExpiration)

	return task, nil
}

func (r *CachedTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	l := logger.FromContext(ctx)
	if err := r.next.Save(ctx, task); err != nil {
		return err
	}

	r.c.Set(task.ID, task, cache.DefaultExpiration)
	l.Info("task saved and cache updated", zap.String("task_id", task.ID))

	return nil
}