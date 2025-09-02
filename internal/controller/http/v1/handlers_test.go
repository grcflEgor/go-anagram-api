package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)


func TestHandlers(t *testing.T) {
	t.Run("WriteError", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := &APIError{
			Code:    "TEST_CODE",
			Message: "test message",
			Details: "details",
			Status:  http.StatusTeapot,
		}
		WriteError(rec, err)

		assert.Equal(t, http.StatusTeapot, rec.Code)

		var resp ErrorResponse
		decodeErr := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, decodeErr)
		assert.Equal(t, "test message", resp.Error)
		assert.Equal(t, "TEST_CODE", resp.Code)
		assert.Equal(t, "details", resp.Details)
	})

	t.Run("APIError_Error", func(t *testing.T) {
		err := &APIError{Message: "test message"}
		assert.Equal(t, "test message", err.Error())
	})

	t.Run("GroupRequest_Validation", func(t *testing.T) {
		gr := GroupRequest{Words: []string{"one"}, CaseSensitive: false}
		b, err := json.Marshal(gr)
		require.NoError(t, err)

		var gr2 GroupRequest
		err = json.Unmarshal(b, &gr2)
		require.NoError(t, err)

		assert.Equal(t, 1, len(gr2.Words))
		assert.Equal(t, "one", gr2.Words[0])
		assert.False(t, gr2.CaseSensitive)
	})

	t.Run("GroupResponse_JSON", func(t *testing.T) {
		resp := GroupResponse{
			TaskID:         "id",
			Status:         "ok",
			Result:         [][]string{{"a", "b"}},
			ProcessingTime: 123,
			GroupsCount:    1,
		}

		b, err := json.Marshal(resp)
		require.NoError(t, err)

		var resp2 GroupResponse
		err = json.Unmarshal(b, &resp2)
		require.NoError(t, err)

		assert.Equal(t, "id", resp2.TaskID)
		assert.Equal(t, "ok", resp2.Status)
		assert.Equal(t, int64(123), resp2.ProcessingTime)
		assert.Equal(t, 1, resp2.GroupsCount)
	})

	t.Run("GroupAnagrams", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockService, _, handlers := setupTestHandlers()
			mockService.On("CreateTask", mock.Anything, []string{"hello", "world"}, false).Return("task123", nil)

			request := GroupRequest{Words: []string{"hello", "world"}, CaseSensitive: false}
			req := createJSONRequest("POST", "/api/v1/anagrams/group", request)
			rec := httptest.NewRecorder()

			handlers.GroupAnagrams(rec, req)

			assert.Equal(t, http.StatusAccepted, rec.Code)

			var response CreateTaskResponse
			err := json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, "task123", response.TaskID)

			mockService.AssertExpectations(t)
		})

		t.Run("InvalidJSON", func(t *testing.T) {
			_, _, handlers := setupTestHandlers()

			req := httptest.NewRequest("POST", "/api/v1/anagrams/group", strings.NewReader(`{"words": ["test", "invalid json`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handlers.GroupAnagrams(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assertErrorResponse(t, rec, "INVALID_REQUEST")
		})

		t.Run("ValidationFailed", func(t *testing.T) {
			_, _, handlers := setupTestHandlers()

			request := GroupRequest{Words: []string{}, CaseSensitive: false}
			req := createJSONRequest("POST", "/api/v1/anagrams/group", request)
			rec := httptest.NewRecorder()

			handlers.GroupAnagrams(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assertErrorResponse(t, rec, "VALIDATION_FAILED")
		})

		t.Run("ServiceError", func(t *testing.T) {
			mockService, _, handlers := setupTestHandlers()
			mockService.On("CreateTask", mock.Anything, []string{"test"}, false).Return("", fmt.Errorf("service error"))

			request := GroupRequest{Words: []string{"test"}, CaseSensitive: false}
			req := createJSONRequest("POST", "/api/v1/anagrams/group", request)
			rec := httptest.NewRecorder()

			handlers.GroupAnagrams(rec, req)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assertErrorResponse(t, rec, "TASK_CREATION_FAILED")

			mockService.AssertExpectations(t)
		})
	})

	t.Run("UploadFile", func(t *testing.T) {
		t.Run("EmptyFileContent", func(t *testing.T) {
			_, _, handlers := setupTestHandlers()

			req := createMultipartRequest("test.txt", "", false)
			rec := httptest.NewRecorder()

			handlers.UploadFile(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assertErrorResponse(t, rec, "INVALID_REQUEST")
		})

		t.Run("SingleWord", func(t *testing.T) {
			mockService, _, handlers := setupTestHandlers()
			mockService.On("CreateTask", mock.Anything, []string{"hello"}, false).Return("task123", nil)

			req := createMultipartRequest("test.txt", "hello", false)
			rec := httptest.NewRecorder()

			handlers.UploadFile(rec, req)

			assert.Equal(t, http.StatusAccepted, rec.Code)
			mockService.AssertExpectations(t)
		})
	})

	t.Run("GetResult", func(t *testing.T) {
		t.Run("MissingTaskID", func(t *testing.T) {
			_, _, handlers := setupTestHandlers()

			req := httptest.NewRequest("GET", "/api/v1/anagrams/groups/", nil)
			rec := httptest.NewRecorder()

			handlers.GetResult(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assertErrorResponse(t, rec, "MISSING_TASK_ID")
		})
	})

	t.Run("HealthCheck", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			_, _, handlers := setupTestHandlers()

			req := httptest.NewRequest("GET", "/health", nil)
			rec := httptest.NewRecorder()

			handlers.HealthCheck(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var response HealthResponse
			err := json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, "ok", response.Status)
		})
	})

	t.Run("GetStats", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			_, stats, handlers := setupTestHandlers()
			stats.On("Get").Return(map[string]uint64{
				"total_tasks":     10,
				"completed_tasks": 8,
				"failed_tasks":    2,
			})

			req := httptest.NewRequest("GET", "/api/v1/anagrams/stats", nil)
			rec := httptest.NewRecorder()

			handlers.GetStats(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var response map[string]uint64
			err := json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, uint64(10), response["total_tasks"])
			assert.Equal(t, uint64(8), response["completed_tasks"])
			assert.Equal(t, uint64(2), response["failed_tasks"])

			stats.AssertExpectations(t)
		})
	})

	t.Run("ClearCache", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockService, _, handlers := setupTestHandlers()
			mockService.On("ClearCache", mock.Anything).Return(nil)

			req := httptest.NewRequest("DELETE", "/api/v1/anagrams/cache", nil)
			rec := httptest.NewRecorder()

			handlers.ClearCache(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)

			mockService.AssertExpectations(t)
		})

		t.Run("ServiceError", func(t *testing.T) {
			mockService, _, handlers := setupTestHandlers()
			mockService.On("ClearCache", mock.Anything).Return(fmt.Errorf("cache clear failed"))

			req := httptest.NewRequest("DELETE", "/api/v1/anagrams/cache", nil)
			rec := httptest.NewRecorder()

			handlers.ClearCache(rec, req)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assertErrorResponse(t, rec, "INTERNAL_SERVER_ERROR")

			mockService.AssertExpectations(t)
		})
	})

	t.Run("Validation", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			validator := validator.New()
			request := GroupRequest{Words: []string{"test", "tset"}, CaseSensitive: true}

			err := validator.Struct(request)
			assert.NoError(t, err)
		})

		t.Run("EmptyWords", func(t *testing.T) {
			validator := validator.New()
			request := GroupRequest{Words: []string{}, CaseSensitive: false}

			err := validator.Struct(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "min")
		})

		t.Run("EmptyStringInWords", func(t *testing.T) {
			validator := validator.New()
			request := GroupRequest{Words: []string{"test", "", "hello"}, CaseSensitive: false}

			err := validator.Struct(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "required")
		})
	})

	t.Run("JSONSerialization", func(t *testing.T) {
		t.Run("GroupRequest", func(t *testing.T) {
			request := GroupRequest{
				Words:         []string{"test", "tset", "привет"},
				CaseSensitive: true,
			}

			data, err := json.Marshal(request)
			require.NoError(t, err)

			var unmarshaled GroupRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, request.Words, unmarshaled.Words)
			assert.Equal(t, request.CaseSensitive, unmarshaled.CaseSensitive)
		})

		t.Run("GroupResponse", func(t *testing.T) {
			response := GroupResponse{
				TaskID:         "task123",
				Status:         "completed",
				Result:         [][]string{{"test", "tset"}},
				ProcessingTime: 100,
				GroupsCount:    1,
			}

			data, err := json.Marshal(response)
			require.NoError(t, err)

			var unmarshaled GroupResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, response.TaskID, unmarshaled.TaskID)
			assert.Equal(t, response.Status, unmarshaled.Status)
			assert.Equal(t, response.ProcessingTime, unmarshaled.ProcessingTime)
			assert.Equal(t, response.GroupsCount, unmarshaled.GroupsCount)
		})
	})
}
