package domain

import "time"

// TaskStatus представляет статус выполнения задачи
type TaskStatus string

const (
	StatusProcessing TaskStatus = "processing" // Задача в обработке
	StatusCompleted  TaskStatus = "completed"  // Задача завершена успешно
	StatusFailed     TaskStatus = "failed"     // Задача завершена с ошибкой
)

// Task представляет задачу по группировке анаграмм
type Task struct {
	// Уникальный идентификатор задачи
	ID string `json:"task_id" example:"task-123"`
	// Статус выполнения задачи
	Status TaskStatus `json:"status" example:"completed"`
	// Список слов для обработки (скрыто из JSON)
	Words []string `json:"-"`
	// Путь к временному файлу (скрыто из JSON)
	FilePath string `json:"-"`
	// Учитывать ли регистр (скрыто из JSON)
	CaseSensitive bool `json:"-"`
	// Результат группировки анаграмм
	Result [][]string `json:"result,omitempty"`
	// Описание ошибки, если задача завершилась неудачно
	Error string `json:"error,omitempty" example:"timeout exceeded"`
	// Время создания задачи (скрыто из JSON)
	CreatedAt time.Time `json:"-"`
	// Время обработки в миллисекундах
	ProcessingTimeMS int64 `json:"processing_time_ms" example:"150"`
	// Количество групп анаграмм
	GroupsCount int `json:"groups_count" example:"2"`

	// Контекст трассировки (скрыто из JSON)
	TraceContext map[string]string `json:"-"`
}
