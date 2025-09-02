package service

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
)

type AnagramServiceProvider interface {
	CreateTask(ctx context.Context, words []string, caseSensitive bool) (string, error)
	GetTaskByID(ctx context.Context, id string) (*domain.Task, error)
	ClearCache(ctx context.Context) error
}


type TaskStatsProvider interface {
	IncrementTotalTasks()
	IncrementCompletedTasks()
	IncrementFailedTasks()
	Get() map[string]uint64
}
