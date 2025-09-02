package v1

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse представляет ответ с ошибкой для клиента
type ErrorResponse struct {
	// Сообщение об ошибке
	Error string `json:"error" example:"invalid request data"`
	// Код ошибки
	Code string `json:"code" example:"INVALID_REQUEST"`
	// Детали ошибки (опционально)
	Details string `json:"details,omitempty" example:"validation failed"`
}

// APIError представляет внутреннюю ошибку API
type APIError struct {
	// Код ошибки
	Code string
	// Сообщение об ошибке
	Message string
	// Детали ошибки
	Details string
	// HTTP статус код
	Status int
}

func (e *APIError) Error() string {
	return e.Message
}

func WriteError(w http.ResponseWriter, err *APIError) {
	response := ErrorResponse{
		Error:   err.Message,
		Code:    err.Code,
		Details: err.Details,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(response)
}

var (
	// ErrInvalidRequest ошибка некорректного запроса
	ErrInvalidRequest = &APIError{
		Code:    "INVALID_REQUEST",
		Message: "invalid request data",
		Status:  http.StatusBadRequest,
	}

	// ErrTaskNotFound ошибка отсутствия задачи
	ErrTaskNotFound = &APIError{
		Code:    "TASK_NOT_FOUND",
		Message: "task not found",
		Status:  http.StatusNotFound,
	}

	// ErrInternalServer внутренняя ошибка сервера
	ErrInternalServer = &APIError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "internal server error",
		Status:  http.StatusInternalServerError,
	}

	// ErrTaskCreationFailed ошибка создания задачи
	ErrTaskCreationFailed = &APIError{
		Code:    "TASK_CREATION_FAILED",
		Message: "failed to create task",
		Status:  http.StatusInternalServerError,
	}
)
