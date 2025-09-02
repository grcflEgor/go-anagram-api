package v1

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/grcflEgor/go-anagram-api/internal/config"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"go.uber.org/zap"
)

type Handlers struct {
	anagramService service.AnagramServiceProvider
	validator      *validator.Validate
	config         *config.Config
	stats          service.TaskStatsProvider
}

func NewHandlers(anagramService service.AnagramServiceProvider, validator *validator.Validate, config *config.Config, stats service.TaskStatsProvider) *Handlers {
	return &Handlers{
		anagramService: anagramService,
		validator:      validator,
		config:         config,
		stats:          stats,
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
			Message: "validation failed",
			Details: err.Error(),
			Status:  http.StatusBadRequest,
		}
		WriteError(w, validationError)
		return
	}

	taskID, err := h.anagramService.CreateTask(r.Context(), request.Words, request.CaseSensitive)
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
			Message: "task ID is required",
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

func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, h.config.Upload.MaxFileSize)
	if err := r.ParseMultipartForm(h.config.Upload.MaxFileSize); err != nil {
		l.Error("failed to parse multipart form", zap.Error(err))
		WriteError(w, ErrInvalidRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		l.Error("failed to get file", zap.Error(err))
		WriteError(w, ErrInvalidRequest)
		return
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		wordsInLine := strings.Fields(line)
		words = append(words, wordsInLine...)
	}

	if err := scanner.Err(); err != nil {
		l.Error("failed to scan file", zap.Error(err))
		WriteError(w, ErrInvalidRequest)
		return
	}

	if len(words) == 0 {
		l.Info("no words found in file")
		WriteError(w, ErrInvalidRequest)
		return
	}

	caseSensitive := false
	if values := r.MultipartForm.Value["case_sensitive"]; len(values) > 0 {
		if strings.ToLower(values[0]) == "true" {
			caseSensitive = true
		}
	}

	taskID, err := h.anagramService.CreateTask(ctx, words, caseSensitive)
	if err != nil {
		l.Error("failed to create task", zap.Error(err))
		WriteError(w, ErrTaskCreationFailed)
		return
	}

	resp := CreateTaskResponse{TaskID: taskID}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		l.Error("failed to write response", zap.Error(err))
	}

}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	stats := h.stats.Get()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		l.Error("failed to write stats", zap.Error(err))
		WriteError(w, ErrInternalServer)
		return
	}
}

type cacheCleaner interface {
	ClearCache(ctx context.Context) error
}

func (h *Handlers) ClearCache(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	if flusher, ok := h.anagramService.(cacheCleaner); ok {
		if err := flusher.ClearCache(r.Context()); err != nil {
			l.Error("failed to clear cache", zap.Error(err))
			WriteError(w, ErrInternalServer)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	WriteError(w, ErrInternalServer)
}
