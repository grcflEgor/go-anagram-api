package storage

import (
	"context"
	"testing"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/test/integration/mocks"
	"github.com/patrickmn/go-cache"
)

func newTestCache(base TaskStorage) *CachedTaskStorage {
	c := cache.New(5*time.Minute, 10*time.Minute)
	return NewCachedTaskStorage(base, c)
}

func TestCachedStorage_SaveAndGet(t *testing.T) {
	base := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := newTestCache(base)

	task := &domain.Task{ID: "1"}
	if err := cache.Save(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := cache.GetByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "1" {
		t.Errorf("expected id=1, got %v", got.ID)
	}
}

func TestCachedStorage_GetByID_CacheHit(t *testing.T) {
	base := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := newTestCache(base)

	task := &domain.Task{ID: "2"}
	cache.cache.Set("2", task, time.Minute)

	got, err := cache.GetByID(context.Background(), "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "2" {
		t.Errorf("expected id=2, got %v", got.ID)
	}
}

func TestCachedStorage_Flush(t *testing.T) {
	base := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task)}
	cache := newTestCache(base)

	task := &domain.Task{ID: "1"}
	cache.cache.Set(task.ID, task, time.Minute)

	err := cache.Flush(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, found := cache.cache.Get(task.ID); found {
		t.Error("expected cache to be flushed, but task still exists")
	}
}
