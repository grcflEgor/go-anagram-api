package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/test/integration/mocks"
)

func TestAnagramService_CreateTask_EdgeCases(t *testing.T) {
	cases := []struct {
		name      string
		words     []string
		batchSize int
		saveErr   error
		wantErr   bool
	}{
		{
			name:      "EmptyWords",
			words:     []string{},
			batchSize: 10,
			wantErr:   false,
		},
		{
			name: "LargeBatch",
			words: func() []string {
				w := make([]string, 20)
				for i := range w {
					w[i] = "w"
				}
				return w
			}(),
			batchSize: 10,
			wantErr:   false,
		},
		{
			name:      "SaveError",
			words:     []string{"one"},
			batchSize: 10,
			saveErr:   errors.New("save error"),
			wantErr:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task), SaveErr: tc.saveErr}
			taskQueue := make(chan *domain.Task, 1)
			stats := NewTaskStats()
			service := NewAnagramService(storage, taskQueue, stats, tc.batchSize)
			ctx := context.Background()
			id, err := service.CreateTask(ctx, tc.words, false)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if id == "" {
					t.Error("expected non-empty id")
				}
			}
		})
	}
}

func TestAnagramService_GetTaskByID_EdgeCases(t *testing.T) {
	cases := []struct {
		name    string
		id      string
		getErr  error
		wantErr bool
	}{
		{
			name:    "NotFound",
			id:      "notfound",
			getErr:  nil,
			wantErr: true,
		},
		{
			name:    "GetError",
			id:      "id1",
			getErr:  errors.New("get error"),
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage := &mocks.MockTaskStorage{Tasks: make(map[string]*domain.Task), GetErr: tc.getErr}
			taskQueue := make(chan *domain.Task, 1)
			stats := NewTaskStats()
			service := NewAnagramService(storage, taskQueue, stats, 10)
			ctx := context.Background()
			if tc.getErr == nil && tc.id != "notfound" {
				storage.Tasks[tc.id] = &domain.Task{ID: tc.id, CreatedAt: time.Now()}
			}
			_, err := service.GetTaskByID(ctx, tc.id)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
