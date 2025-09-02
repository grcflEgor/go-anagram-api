package mocks

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockAnagramService struct {
	mock.Mock
}

func (m *MockAnagramService) CreateTask(ctx context.Context, words []string, caseSensitive bool) (string, error) {
	args := m.Called(ctx, words, caseSensitive)
	return args.String(0), args.Error(1)
}

func (m *MockAnagramService) GetTaskByID(ctx context.Context, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockAnagramService) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}