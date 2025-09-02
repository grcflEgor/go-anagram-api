package v1

// GroupResponse представляет ответ с результатом группировки анаграмм
type GroupResponse struct {
	// Уникальный идентификатор задачи
	TaskID string `json:"task_id" example:"task-123"`
	// Статус выполнения задачи
	Status string `json:"status" example:"completed"`
	// Результат группировки анаграмм
	Result [][]string `json:"result"`
	// Время обработки в миллисекундах
	ProcessingTime int64 `json:"processing_time_ms" example:"150"`
	// Количество групп анаграмм
	GroupsCount int `json:"groups_count" example:"2"`
}

// CreateTaskResponse представляет ответ при создании задачи
type CreateTaskResponse struct {
	// Уникальный идентификатор созданной задачи
	TaskID string `json:"task_id" example:"task-123"`
}

// HealthResponse представляет ответ проверки здоровья сервиса
type HealthResponse struct {
	// Статус сервиса
	Status string `json:"status" example:"ok"`
}

// StatsResponse представляет статистику по задачам
type StatsResponse struct {
	// Общее количество задач
	TotalTasks int64 `json:"total_tasks" example:"100"`
	// Количество завершенных задач
	CompletedTasks int64 `json:"completed_tasks" example:"85"`
	// Количество неудачных задач
	FailedTasks int64 `json:"failed_tasks" example:"5"`
}
