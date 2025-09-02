package service

import "sync/atomic"

type TaskStats struct {
	TotalTasks atomic.Uint64
	CompletedTasks atomic.Uint64
	FailedTasks atomic.Uint64
}

func NewTaskStats() *TaskStats {
	return &TaskStats{}
}

func (ts *TaskStats) IncrementTotalTasks() {
	ts.TotalTasks.Add(1)
}

func (ts *TaskStats) IncrementCompletedTasks() {
	ts.CompletedTasks.Add(1)
}

func (ts *TaskStats) IncrementFailedTasks() {
	ts.FailedTasks.Add(1)
}


func (ts *TaskStats) Get() map[string]uint64 {
	return map[string]uint64{
		"total_tasks": ts.TotalTasks.Load(),
		"completed_tasks": ts.CompletedTasks.Load(),
		"failed_tasks": ts.FailedTasks.Load(),
	}
}