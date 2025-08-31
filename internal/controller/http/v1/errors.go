package v1

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

type APIError struct {
	Code    string
	Message string
	Details string
	Status  int
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
	ErrInvalidRequest = &APIError{
		Code:    "INVALID_REQUEST",
		Message: "invalid request data",
		Status:  http.StatusBadRequest,
	}

	ErrTaskNotFound = &APIError{
		Code:    "TASK_NOT_FOUND",
		Message: "task not found",
		Status:  http.StatusNotFound,
	}

	ErrInternalServer = &APIError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "internal server error",
		Status:  http.StatusInternalServerError,
	}

	ErrTaskCreationFailed = &APIError{
		Code:    "TASK_CREATION_FAILED",
		Message: "failed to create task",
		Status:  http.StatusInternalServerError,
	}
)
