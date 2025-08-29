package domain

import "time"

type TaskStatus string

const (
	StatusProcessing TaskStatus = "processing"
	StatusCompleted TaskStatus = "completed"
	StatusFailed TaskStatus = "failed"
)

type Task struct {
	ID string `json:"task_id"`
	Status TaskStatus `json:"status"`
	Words []string `json:"-"`
	Result [][]string `json:"result,omitempty"`
	Error string `json:"error,omitempty"`
	CreatedAt time.Time `json:"-"`
	ProcessingTimeMS int64 `json:"processing_time_ms"`
	GroupsCount int `json:"groups_count"`

	TraceContext map[string]string `json:"-"`
}