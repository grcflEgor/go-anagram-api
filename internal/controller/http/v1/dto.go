package v1

type GroupRequest struct{
	Words []string `json:"words" validate:"min=1,dive,required"`
}

type GroupResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
	Result [][]string `json:"result"`
	ProcessingTime int64 `json:"processing_time_ms"`
	GroupsCount int `json:"groups_count"`
}

type CreateTaskResponse struct {
	TaskID string `json:"task_id"`
}