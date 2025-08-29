package v1

type GroupResponse struct {
	TaskID         string     `json:"task_id"`
	Status         string     `json:"status"`
	Result         [][]string `json:"result"`
	ProcessingTime int64      `json:"processing_time_ms"`
	GroupsCount    int        `json:"groups_count"`
}

type CreateTaskResponse struct {
	TaskID string `json:"task_id"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type StatsResponse struct {
	TotalTasks     int64 `json:"total_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`
	CacheHits      int64 `json:"cache_hits"`
	CacheMisses    int64 `json:"cache_misses"`
}
