package mocks

import (
	"context"
	"errors"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
)


type MockTaskStorage struct {
	Tasks map[string]*domain.Task
	SaveErr error
	GetErr error
	FlushFn func(ctx context.Context) error
}

func (m *MockTaskStorage) Save(ctx context.Context, task *domain.Task) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	if m.Tasks == nil {
		m.Tasks = make(map[string]*domain.Task)
	}
	m.Tasks[task.ID] = task
	return nil
}

func (m *MockTaskStorage) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	if m.Tasks == nil {
		return nil, errors.New("not found")
	}
	t, ok := m.Tasks[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (m *MockTaskStorage) Flush(ctx context.Context) error {
	if m.FlushFn != nil {
		return m.FlushFn(ctx)
	}
	m.Tasks = make(map[string]*domain.Task)
	return nil
}