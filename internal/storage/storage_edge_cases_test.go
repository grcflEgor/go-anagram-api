package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/test/integration/mocks"
)

func TestCachedStorage_GetByID_BaseError(t *testing.T) {
	base := &mocks.MockTaskStorage{GetErr: errors.New("db error")}
	cache := newTestCache(base)

	_, err := cache.GetByID(context.Background(), "not-exist")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInMemoryStorage_GetByID_NotFound(t *testing.T) {
	store := NewInMemoryStorage()
	_, err := store.GetByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMockTaskStorage_SaveErr(t *testing.T) {
	store := &mocks.MockTaskStorage{SaveErr: errors.New("save failed")}
	err := store.Save(context.Background(), &domain.Task{ID: "1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := store.Tasks["1"]; ok {
		t.Fatal("task should not be saved when SaveErr is set")
	}
}

func TestMockTaskStorage_GetErr(t *testing.T) {
	store := &mocks.MockTaskStorage{GetErr: errors.New("get failed")}
	_, err := store.GetByID(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMockTaskStorage_FlushFn(t *testing.T) {
	called := false
	store := &mocks.MockTaskStorage{
		FlushFn: func(ctx context.Context) error {
			called = true
			return nil
		},
	}
	err := store.Flush(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected FlushFn to be called")
	}
}

func TestMockTaskStorage_Flush_Default(t *testing.T) {
	store := &mocks.MockTaskStorage{Tasks: map[string]*domain.Task{"1": {ID: "1"}}}
	_ = store.Flush(context.Background())
	if len(store.Tasks) != 0 {
		t.Fatal("expected tasks to be cleared on default Flush")
	}
}
