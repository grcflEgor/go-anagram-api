package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/internal/test/integration/mocks"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestMerge(t *testing.T) {
	target := map[string][]string{
		"a": {"a1"},
	}
	source := map[string][]string{
		"a": {"a2"},
		"b": {"b1"},
	}
	merge(target, source)

	if len(target["a"]) != 2 {
		t.Errorf("expected 2 items for key a, got %d", len(target["a"]))
	}
	if len(target["b"]) != 1 {
		t.Errorf("expected 1 item for key b, got %d", len(target["b"]))
	}
}

func TestStop(t *testing.T) {
	storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := &mocks.CacheTaskStorage{}
	taskQueue := make(chan *domain.Task, 1)
	logger := zap.NewNop()
	stats := service.NewTaskStats()

	pool := NewPool(storage, cache, taskQueue, logger, time.Second, stats, 10)
	go pool.Run(1)

	pool.Stop()
}

func TestWorker_ProcessWordsTask_Success(t *testing.T) {
	storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := &mocks.CacheTaskStorage{}
	cache.On("Save", mock.Anything, mock.Anything).Return(nil)
	taskQueue := make(chan *domain.Task, 1)
	logger := zap.NewNop()
	stats := service.NewTaskStats()

	pool := NewPool(storage, cache, taskQueue, logger, time.Second, stats, 10)
	go pool.Run(1)
	defer pool.Stop()

	task := &domain.Task{
		ID:            "t1",
		Words:         []string{"кот", "ток"},
		CaseSensitive: false,
		TraceContext:  make(map[string]string),
	}
	taskQueue <- task

	time.Sleep(200 * time.Millisecond)

	saved, _ := storage.GetByID(context.Background(), "t1")
	if saved.Status != domain.StatusCompleted {
		t.Errorf("expected Completed, got %v", saved.Status)
	}
	if saved.GroupsCount == 0 {
		t.Errorf("expected groups > 0, got %d", saved.GroupsCount)
	}
	if stats.CompletedTasks.Load() != 1 {
		t.Errorf("expected 1 completed, got %d", stats.CompletedTasks.Load())
	}
}

func TestWorker_ProcessFileTask_Error(t *testing.T) {
	storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := &mocks.CacheTaskStorage{}
	cache.On("Save", mock.Anything, mock.Anything).Return(nil)
	taskQueue := make(chan *domain.Task, 1)
	logger := zap.NewNop()
	stats := service.NewTaskStats()

	pool := NewPool(storage, cache, taskQueue, logger, time.Second, stats, 2)
	go pool.Run(1)
	defer pool.Stop()

	task := &domain.Task{
		ID:       "t2",
		FilePath: "not_exists.txt",
	}
	taskQueue <- task

	time.Sleep(200 * time.Millisecond)

	saved, _ := storage.GetByID(context.Background(), "t2")
	if saved.Status != domain.StatusFailed {
		t.Errorf("expected Failed, got %v", saved.Status)
	}
	if stats.FailedTasks.Load() != 1 {
		t.Errorf("expected 1 failed, got %d", stats.FailedTasks.Load())
	}
}

func TestWorker_ProcessFileTask_Success(t *testing.T) {
	storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := &mocks.CacheTaskStorage{}
	cache.On("Save", mock.Anything, mock.Anything).Return(nil)
	taskQueue := make(chan *domain.Task, 1)
	logger := zap.NewNop()
	stats := service.NewTaskStats()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "words.txt")
	err := os.WriteFile(filePath, []byte("кот\nток\nрост\nторс\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	pool := NewPool(storage, cache, taskQueue, logger, time.Second, stats, 2)
	go pool.Run(1)
	defer pool.Stop()

	task := &domain.Task{
		ID:       "t3",
		FilePath: filePath,
	}
	taskQueue <- task

	time.Sleep(300 * time.Millisecond)

	saved, _ := storage.GetByID(context.Background(), "t3")
	if saved.Status != domain.StatusCompleted {
		t.Errorf("expected Completed, got %v", saved.Status)
	}
	if saved.GroupsCount == 0 {
		t.Errorf("expected groups > 0, got %d", saved.GroupsCount)
	}
	if stats.CompletedTasks.Load() != 1 {
		t.Errorf("expected 1 completed, got %d", stats.CompletedTasks.Load())
	}
}

func TestWorker_TaskTimeout(t *testing.T) {
	t.Parallel()
	storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := &mocks.CacheTaskStorage{}
	cache.On("Save", mock.Anything, mock.Anything).Return(nil)
	taskQueue := make(chan *domain.Task, 1)
	logger := zap.NewNop()
	stats := service.NewTaskStats()

	pool := NewPool(storage, cache, taskQueue, logger, 1*time.Nanosecond, stats, 10)
	go pool.Run(1)
	defer pool.Stop()

	task := &domain.Task{
		ID:            "t4",
		Words:         []string{"кот", "ток"},
		CaseSensitive: false,
		TraceContext:  make(map[string]string),
	}
	taskQueue <- task

	time.Sleep(200 * time.Millisecond)

	saved, _ := storage.GetByID(context.Background(), "t4")
	if saved.Status != domain.StatusFailed {
		t.Errorf("expected Failed due to timeout, got %v", saved.Status)
	}
	if stats.FailedTasks.Load() != 1 {
		t.Errorf("expected 1 failed, got %d", stats.FailedTasks.Load())
	}
}
