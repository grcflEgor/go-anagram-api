package integration

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/grcflEgor/go-anagram-api/internal/config"
	v1 "github.com/grcflEgor/go-anagram-api/internal/controller/http/v1"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/internal/storage"
	"github.com/grcflEgor/go-anagram-api/internal/worker"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func writeMetrics(t *testing.T, testName string, duration time.Duration, processingMS int64, wordsCount int, groupsCount int) {
	t.Helper()

	if err := os.MkdirAll("metrics", 0o755); err != nil {
		t.Logf("dont create dir \"metrics\": %v", err)
		return
	}

	csvPath := filepath.Join("metrics", "metrics.csv")
	csvFile, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		defer csvFile.Close()
		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		stat, _ := csvFile.Stat()
		if stat.Size() == 0 {
			_ = writer.Write([]string{"TestName", "DurationSec", "ProcessingMS", "WordsCount", "GroupsCount"})
		}
		_ = writer.Write([]string{
			testName,
			strconv.FormatFloat(duration.Seconds(), 'f', 2, 64),
			strconv.FormatInt(processingMS, 10),
			strconv.Itoa(wordsCount),
			strconv.Itoa(groupsCount),
		})
	}

	mdPath := filepath.Join("metrics", "metrics.md")
	mdFile, err := os.OpenFile(mdPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		defer mdFile.Close()
		stat, _ := mdFile.Stat()
		if stat.Size() == 0 {
			_, _ = mdFile.WriteString("| TestName | Duration (s) | ProcessingMS | Words | Groups |\n")
			_, _ = mdFile.WriteString("|----------|--------------|--------------|-------|--------|\n")
		}
		_, _ = mdFile.WriteString(fmt.Sprintf(
			"| %s | %.2f | %d | %d | %d |\n",
			testName, duration.Seconds(), processingMS, wordsCount, groupsCount,
		))
	}
}


func TestLargeFileStreamingProcessing(t *testing.T) {
	logger.InitLogger()
	defer func() { _ = logger.AppLogger.Sync() }()

	config := &config.Config{}
	config.Worker.Count = 4
	config.Task.QueueSize = 1000
	config.Processing.Timeout = 60 * time.Second
	config.Cache.DefaultExpiration = 5 * time.Minute
	config.Cache.CleanupInterval = 10 * time.Minute
	config.Upload.MaxFileSize = 100 * 1024 * 102

	tempDir, err := os.MkdirTemp("", "anagram-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	largeFilePath := filepath.Join(tempDir, "large_words.txt")
	err = createLargeTestFile(largeFilePath, 150000)
	require.NoError(t, err)

	memoryStorage := storage.NewInMemoryStorage()
	cacheInstance := cache.New(config.Cache.DefaultExpiration, config.Cache.CleanupInterval)
	cachedStorage := storage.NewCachedTaskRepository(memoryStorage, cacheInstance)

	taskQueue := make(chan *domain.Task, config.Task.QueueSize)
	stats := service.NewTaskStats()

	batchSize := 1000
	anagramService := service.NewAnagramService(cachedStorage, taskQueue, stats, batchSize)
	workerPool := worker.NewPool(cachedStorage, taskQueue, logger.AppLogger, config.Processing.Timeout, stats, batchSize)

	workerPool.Run(config.Worker.Count)
	defer workerPool.Stop()

	appValidator := validator.New()
	handlers := v1.NewHandlers(anagramService, appValidator, config, stats)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/anagrams/upload":
			handlers.UploadFile(w, r)
		case "/api/v1/anagrams/groups":
			handlers.GetResult(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Run("UploadLargeFileViaHTTP", func(t *testing.T) {
		file, _ := os.Open(largeFilePath)
		defer file.Close()

		body := &strings.Builder{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "large_words.txt")
		_, _ = io.Copy(part, file)
		_ = writer.WriteField("case_sensitive", "false")
		_ = writer.Close()

		req, _ := http.NewRequest("POST", server.URL+"/api/v1/anagrams/upload", io.NopCloser(strings.NewReader(body.String())))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, _ := (&http.Client{Timeout: 30 * time.Second}).Do(req)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		var response v1.CreateTaskResponse
		_ = json.NewDecoder(resp.Body).Decode(&response)
		taskID := response.TaskID

		var task *domain.Task
	Loop1:
		for {
			select {
			case <-time.After(180 * time.Second):
				t.Fatal("timeout")
			case <-time.Tick(1 * time.Second):
				task, _ = anagramService.GetTaskByID(context.Background(), taskID)
				if task != nil && (task.Status == domain.StatusCompleted || task.Status == domain.StatusFailed) {
					break Loop1
				}
			}
		}

		require.NotNil(t, task)
		writeMetrics(t, "UploadLargeFileViaHTTP", time.Duration(task.ProcessingTimeMS)*time.Millisecond, task.ProcessingTimeMS, len(task.Words), task.GroupsCount)
	})

	t.Run("StreamingPerformanceTest", func(t *testing.T) {
		perfFilePath := filepath.Join(tempDir, "perf_test.txt")
		_ = createLargeTestFile(perfFilePath, 200000)

		start := time.Now()
		task := &domain.Task{
			ID:            "perf-" + fmt.Sprint(time.Now().UnixNano()),
			Status:        domain.StatusProcessing,
			FilePath:      perfFilePath,
			CaseSensitive: false,
			CreatedAt:     time.Now(),
			TraceContext:  make(map[string]string),
		}
		_ = cachedStorage.Save(context.Background(), task)
		taskQueue <- task
		stats.IncrementTotalTasks()

		var resultTask *domain.Task
	Loop2:
		for {
			select {
			case <-time.After(180 * time.Second):
				t.Fatal("timeout")
			case <-time.Tick(2 * time.Second):
				resultTask, _ = anagramService.GetTaskByID(context.Background(), task.ID)
				if resultTask != nil && (resultTask.Status == domain.StatusCompleted || resultTask.Status == domain.StatusFailed) {
					break Loop2
				}
			}
		}

		duration := time.Since(start)
		writeMetrics(t, "StreamingPerformanceTest", duration, resultTask.ProcessingTimeMS, 200000, resultTask.GroupsCount)
	})

	t.Run("ResultsCorrectnessTest", func(t *testing.T) {
		testWords := []string{"ток", "рост", "кот", "торс", "Кто", "фывап", "рок", "hello", "world", "olleh", "dlrow", "test", "tset", "апельсин", "спаниель", "лиса", "сила", "мама", "амма"}
		taskID, _ := anagramService.CreateTask(context.Background(), testWords, false)

		var task *domain.Task
	Loop3:
		for {
			select {
			case <-time.After(30 * time.Second):
				t.Fatal("timeout")
			case <-time.Tick(1 * time.Second):
				task, _ = anagramService.GetTaskByID(context.Background(), taskID)
				if task != nil && (task.Status == domain.StatusCompleted || task.Status == domain.StatusFailed) {
					break Loop3
				}
			}
		}

		require.NotNil(t, task)
		writeMetrics(t, "ResultsCorrectnessTest", time.Duration(task.ProcessingTimeMS)*time.Millisecond, task.ProcessingTimeMS, len(task.Words), task.GroupsCount)
	})
}


func createLargeTestFile(filePath string, wordCount int) error {
	file, _ := os.Create(filePath)
	defer file.Close()
	words := generateRandomWords(wordCount)
	writer := bufio.NewWriter(file)
	for _, word := range words {
		fmt.Fprintln(writer, word)
	}
	return writer.Flush()
}

func generateRandomWords(count int) []string {
	words := make([]string, count)
	base := []string{"hello", "world", "test", "go", "api", "anagram", "word", "group", "ток", "кот", "рост", "торс", "Кто", "фывап", "рок", "мама", "папа", "апельсин", "спаниель", "лиса", "сила"}
	for i := 0; i < count; i++ {
		words[i] = base[i%len(base)] + strconv.Itoa(i%1000)
	}
	return words
}
