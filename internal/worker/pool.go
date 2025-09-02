package worker

import (
	"bufio"
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/internal/storage"
	"github.com/grcflEgor/go-anagram-api/pkg/anagram"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

type Pool struct {
	storage   storage.TaskStorage
	taskQueue chan *domain.Task
	logger    *zap.Logger
	wg        sync.WaitGroup
	processingTimeout time.Duration
	stats *service.TaskStats
	batchSize int
}

func NewPool(storage storage.TaskStorage, taskQueue chan *domain.Task, logger *zap.Logger, processingTimeout time.Duration, stats *service.TaskStats, batchSize int) *Pool {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	return &Pool{
		storage:   storage,
		taskQueue: taskQueue,
		logger:    logger,
		processingTimeout: processingTimeout,
		stats: stats,
		batchSize: batchSize,
	}
}

func (pool *Pool) Run(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker(i + 1)
	}
}

func (pool *Pool) worker(id int) {
	defer pool.wg.Done()

	workerLog := pool.logger.With(zap.Int("worker_id", id))
	workerLog.Info("worker started")
	tr := otel.Tracer("worker")

	for task := range pool.taskQueue {
		func(task *domain.Task) {
			parentCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(task.TraceContext))

			taskCtx, cancel := context.WithTimeout(parentCtx, pool.processingTimeout)
			defer cancel()

			spanCtx, span := tr.Start(taskCtx, "process_task")
			defer span.End()

			span.SetAttributes(attribute.String("task_id", task.ID))
			taskLog := workerLog.With(zap.String("task_id", task.ID))
			taskLog.Info("processing task")

			start := time.Now()

			var grouped map[string][]string
			var err error 

			if task.FilePath != "" {
				grouped, err = pool.processFile(spanCtx, task.FilePath, task.CaseSensitive)
				if removeErr := os.Remove(task.FilePath); removeErr != nil {
					workerLog.Warn("failed to remove file", zap.Error(removeErr))
				}
			} else {
				grouped, err = anagram.Group(spanCtx, task.Words, task.CaseSensitive)
			}

			processingTime := time.Since(start).Milliseconds()
			span.SetAttributes(attribute.Int64("processing_ms", processingTime))

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					taskLog.Warn("task processing timeout")
					task.Status = domain.StatusFailed
					task.Error = "task processing timeout"
					pool.stats.IncrementFailedTasks()
				} else {
					taskLog.Error("task processing failed", zap.Error(err))
					task.Status = domain.StatusFailed
					task.Error = err.Error()
					pool.stats.IncrementFailedTasks()
				}
				span.RecordError(err)
				span.SetAttributes(attribute.String("status", "failed"))
			} else {
				result := make([][]string, 0, len(grouped))
				for _, group := range grouped {
					if len(group) > 1 {
						result = append(result, group)
					}
				}
				task.Status = domain.StatusCompleted
				task.Result = result
				task.ProcessingTimeMS = processingTime
				task.GroupsCount = len(result)
				pool.stats.IncrementCompletedTasks()
				
				span.SetAttributes(attribute.String("status", "completed"))
				span.SetAttributes(attribute.Int("groups_count", len(result)))
			}

			if err := pool.storage.Save(context.Background(), task); err != nil {
				taskLog.Error("failed to save completed task", zap.String("status", string(task.Status)), zap.Error(err))
				span.RecordError(err)
			}
			taskLog.Info("finished task")
		}(task)
	}
}

func (pool *Pool) Stop() {
	close(pool.taskQueue)
	pool.wg.Wait()
}


func (pool *Pool) processFile(ctx context.Context, filePath string, caseSensitive bool) (map[string][]string, error) {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "processFile")
	defer span.End()

	l.Info("processing file", zap.String("file_path", filePath))
	span.SetAttributes(attribute.String("file_path", filePath))
	
	file, err := os.Open(filePath)
	if err != nil {
		l.Error("failed to open file", zap.Error(err))
		span.RecordError(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	batchSize := pool.batchSize
	groups := make(map[string][]string)
	var batch []string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		word := scanner.Text()
		batch = append(batch, word)

		if len(batch) >= batchSize {
			part, err := anagram.Group(ctx, batch, caseSensitive)
			if err != nil {
				return nil, err
			}
			merge(groups, part)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		part, err := anagram.Group(ctx, batch, caseSensitive)
		if err != nil {
			return nil, err
		}
		merge(groups, part)
	}
	return groups, scanner.Err()
}

func merge(target, source map[string][]string) {
	for k, v := range source {
		target[k] = append(target[k], v...)
	}
}