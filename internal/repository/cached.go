package repository

import (
	"context"
	"log"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/patrickmn/go-cache"
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
	if task, found := r.c.Get(id); found {
		log.Println("cache HIT for task", id)
		return task.(*domain.Task), nil
	}
	log.Println("cache MISS for task", id)

	task, err := r.next.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.c.Set(id, task, cache.DefaultExpiration)

	return task, nil
}

func (r *CachedTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	if err := r.next.Save(ctx, task); err != nil {
		return err
	}

	r.c.Set(task.ID, task, cache.DefaultExpiration)
	log.Printf("task %s saved and cache updated", task.ID)

	return nil
}