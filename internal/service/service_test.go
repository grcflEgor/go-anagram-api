package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/domain/repositories"
	"github.com/grcflEgor/go-anagram-api/internal/test/integration/mocks"
	"github.com/stretchr/testify/mock"
)

func TestNewTaskStats(t *testing.T) {
	ts := NewTaskStats()
	if ts == nil {
		t.Fatal("NewTaskStats returned nil")
	}
}

func TestTaskStats_IncrementAndGet(t *testing.T) {
	ts := NewTaskStats()
	ts.IncrementTotalTasks()
	ts.IncrementCompletedTasks()
	ts.IncrementFailedTasks()
	stats := ts.Get()
	if stats["total_tasks"] != 1 || stats["completed_tasks"] != 1 || stats["failed_tasks"] != 1 {
		t.Error("TaskStats counters not incremented correctly")
	}
}

func TestAnagramService_CreateTask(t *testing.T) {
	storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := &mocks.CacheTaskStorage{}
	cache.On("Save", mock.Anything, mock.Anything).Return(nil)
	taskQueue := make(chan *domain.Task, 1)
	stats := NewTaskStats()

	service := NewAnagramService(storage, cache, taskQueue, stats, 10)
	ctx := context.Background()

	words := []string{"one", "two"}
	id, err := service.CreateTask(ctx, words, false)
	if err != nil {
		t.Fatalf("CreateTask error: %v", err)
	}
	if id == "" {
		t.Error("CreateTask returned empty id")
	}
}

func TestAnagramService_GetTaskByID(t *testing.T) {
	storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := &mocks.CacheTaskStorage{}
	cache.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("cache miss"))
	cache.On("Save", mock.Anything, mock.Anything).Return(nil)
	taskQueue := make(chan *domain.Task, 1)
	stats := NewTaskStats()

	service := NewAnagramService(storage, cache, taskQueue, stats, 10)
	ctx := context.Background()

	task := &domain.Task{ID: "id1", CreatedAt: time.Now()}
	storage.Tasks[task.ID] = task

	got, err := service.GetTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTaskByID error: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("GetTaskByID = %v, want %v", got.ID, task.ID)
	}
}

type flusherMock struct {
	called bool
	mock.Mock
}

func (f *flusherMock) Save(ctx context.Context, task *domain.Task) error {
	args := f.Called(ctx, task)
	return args.Error(0)
}
func (f *flusherMock) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	args := f.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}
func (f *flusherMock) Flush(ctx context.Context) error {
	f.called = true
	return nil
}

func (f *flusherMock) Delete(ctx context.Context, id string) error {
	args := f.Called(ctx, id)
	return args.Error(0)
}

func (f *flusherMock) DeleteAll(ctx context.Context) error {
	args := f.Called(ctx)
	return args.Error(0)
}

var _ repositories.CacheTaskStorage = &flusherMock{}

func TestAnagramService_ClearCache(t *testing.T) {
	storage := &mocks.MockTaskStorage{}
	cache := &flusherMock{}
	cache.On("DeleteAll", mock.Anything).Return(nil)
	taskQueue := make(chan *domain.Task, 1)
	stats := NewTaskStats()

	service := NewAnagramService(storage, cache, taskQueue, stats, 10)
	ctx := context.Background()

	err := service.ClearCache(ctx)
	if err != nil {
		t.Errorf("ClearCache error: %v", err)
	}
	cache.AssertCalled(t, "DeleteAll", mock.Anything)
}
