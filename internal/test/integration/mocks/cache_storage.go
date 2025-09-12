package mocks

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/stretchr/testify/mock"
)

type CacheTaskStorage struct {
	mock.Mock
}

func (m *CacheTaskStorage) Save(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *CacheTaskStorage) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *CacheTaskStorage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *CacheTaskStorage) DeleteAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
