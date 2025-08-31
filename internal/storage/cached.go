package storage

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var _ TaskStorage = (*CachedTaskRepository)(nil)

type CachedTaskRepository struct {
	next  TaskStorage
	cache *cache.Cache
}

func NewCachedTaskRepository(next TaskStorage, cache *cache.Cache) *CachedTaskRepository {
	return &CachedTaskRepository{
		next:  next,
		cache: cache,
	}
}

func (r *CachedTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("repository")
	ctx, span := tr.Start(ctx, "CachedTaskRepository.GetByID")
	defer span.End()

	if task, found := r.cache.Get(id); found {
		l.Info("cache HIT for task", zap.String("task_id", id))
		span.SetAttributes(attribute.String("cache", "HIT"))
		return task.(*domain.Task), nil
	}
	span.SetAttributes(attribute.String("cache", "MISS"))
	l.Info("cache MISS for task", zap.String("task_id", id))

	task, err := r.next.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	r.cache.Set(id, task, cache.DefaultExpiration)

	return task, nil
}

func (r *CachedTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("repository")
	ctx, span := tr.Start(ctx, "CachedTaskRepository.Save")
	defer span.End()

	if err := r.next.Save(ctx, task); err != nil {
		span.RecordError(err)
		return err
	}

	r.cache.Set(task.ID, task, cache.DefaultExpiration)
	l.Info("task saved and cache updated", zap.String("task_id", task.ID))

	return nil
}

func (r *CachedTaskRepository) Flush(ctx context.Context) error {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("repository")
	_, span := tr.Start(ctx, "CachedTaskRepository.Flush")
	defer span.End()

	r.cache.Flush()

	l.Info("cache flushed")

	return nil
}