package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"go.uber.org/zap"
)

type Handlers struct {
	anagramService service.AnagramServiceProvider
	validator      *validator.Validate
}

func NewHandlers(anagramService service.AnagramServiceProvider, validator *validator.Validate) *Handlers {
	return &Handlers{
		anagramService: anagramService,
		validator:      validator,
	}
}

func (h *Handlers) GroupAnagrams(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	var request GroupRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		l.Info("invalid request body")
		WriteError(w, ErrInvalidRequest)
		return
	}

	if err := h.validator.Struct(request); err != nil {
		l.Info("validation failed", zap.Error(err))
		validationError := &APIError{
			Code:    "VALIDATION_FAILED",
			Message: "Validation failed",
			Details: err.Error(),
			Status:  http.StatusBadRequest,
		}
		WriteError(w, validationError)
		return
	}

	taskID, err := h.anagramService.CreateTask(r.Context(), request.Words)
	if err != nil {
		l.Error("failed to create task", zap.Error(err))
		WriteError(w, ErrTaskCreationFailed)
		return
	}

	response := CreateTaskResponse{TaskID: taskID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		l.Error("failed to write response", zap.Error(err))
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	response := HealthResponse{Status: "ok"}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		l.Error("failed to write healthcheck response", zap.Error(err))
	}
}

func (h *Handlers) GetResult(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		l.Info("task ID is required")
		missingIDError := &APIError{
			Code:    "MISSING_TASK_ID",
			Message: "Task ID is required",
			Status:  http.StatusBadRequest,
		}
		WriteError(w, missingIDError)
		return
	}

	task, err := h.anagramService.GetTaskByID(r.Context(), taskID)
	if err != nil {
		l.Error("failed to get task by id", zap.Error(err))
		WriteError(w, ErrTaskNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		l.Error("failed to write get result", zap.Error(err))
	}
}
