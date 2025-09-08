package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/domain/repositories"
)

var _ repositories.TaskStorage = (*InMemoryStorage)(nil)

type InMemoryStorage struct {
	mu    sync.RWMutex
	tasks map[string]*domain.Task
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tasks: make(map[string]*domain.Task),
	}
}

func (r *InMemoryStorage) Save(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryStorage) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task with id %s not found", id)
	}
	return task, nil
}
