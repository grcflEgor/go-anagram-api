package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grcflEgor/go-anagram-api/internal/usecase"
)


type GroupRequest struct{
	Words []string `json:"words"`
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

type Handlers struct {
	useCase usecase.AnagramUseCaseProvider
}

func NewHandlers(uc usecase.AnagramUseCaseProvider) *Handlers {
	return &Handlers{useCase: uc}
}

func (h *Handlers) GroupAnagrams(w http.ResponseWriter, r *http.Request) {
	var req GroupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid req body", http.StatusBadRequest)
		return
	}

	taskID, err := h.useCase.CreateTask(r.Context(), req.Words)
	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	resp := CreateTaskResponse{TaskID: taskID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write healthcheck response: %v", err)
	}
}

func (h *Handlers) GetResult(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "task ID is required", http.StatusBadRequest)
		return
	}

	task, err := h.useCase.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to write get result response: %v", err)
	}
}