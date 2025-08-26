package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/grcflEgor/go-anagram-api/pkg/anagram"
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

func GroupAnagramsHandler(w http.ResponseWriter, r *http.Request) {
	var req GroupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid req body", http.StatusBadRequest)
		return
	}

	start := time.Now()
	grouped := anagram.Group(req.Words)
	processingTime := time.Since(start).Milliseconds()

	result := make([][]string, 0, len(grouped))
	for _, group := range grouped {
		if len(group) > 1 {
			result = append(result, group)
		}
	}

	resp := GroupResponse{
		TaskID: uuid.New().String(),
		Status: "completed",
		Result: result,
		ProcessingTime: processingTime,
		GroupsCount: len(result),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "invalid resp body", http.StatusInternalServerError)
		return
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "invalid resp body", http.StatusInternalServerError)
	}
}

func GetResultHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}