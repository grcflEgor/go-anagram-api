package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/grcflEgor/go-anagram-api/internal/config"
	"github.com/grcflEgor/go-anagram-api/internal/test/integration/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandlers() (*mocks.MockAnagramService, *mocks.MockTaskStats, *Handlers) {
	mockService := &mocks.MockAnagramService{}
	validator := validator.New()
	config := &config.Config{}
	config.Upload.MaxFileSize = 100 * 1024 * 1024
	stats := &mocks.MockTaskStats{}
	handlers := NewHandlers(mockService, validator, config, stats)
	return mockService, stats, handlers
}

func createJSONRequest(method, url string, body interface{}) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonData, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonData)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func createMultipartRequest(filename, content string, caseSensitive bool) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		panic(err)
	}
	_, err = part.Write([]byte(content))
	if err != nil {
		panic(err)
	}

	caseSensitivePart, err := writer.CreateFormField("case_sensitive")
	if err != nil {
		panic(err)
	}
	_, err = caseSensitivePart.Write([]byte(fmt.Sprintf("%t", caseSensitive)))
	if err != nil {
		panic(err)
	}

	err = writer.Close()
	if err != nil {
		panic(err)
	}

	req := httptest.NewRequest("POST", "/api/v1/anagrams/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedCode string) {
	var errorResp ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Equal(t, expectedCode, errorResp.Code)
}
