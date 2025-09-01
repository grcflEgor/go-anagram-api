package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/service"
	"github.com/grcflEgor/go-anagram-api/internal/storage"
	"github.com/grcflEgor/go-anagram-api/pkg/anagram"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

const ProcessingTimeout = 30 * time.Second

type Pool struct {
	storage   storage.TaskStorage
	taskQueue chan *domain.Task
	logger    *zap.Logger
	wg        sync.WaitGroup
	processingTimeout time.Duration
	stats *service.TaskStats
}

func NewPool(storage storage.TaskStorage, taskQueue chan *domain.Task, logger *zap.Logger, processingTimeout time.Duration, stats *service.TaskStats) *Pool {
	return &Pool{
		storage:   storage,
		taskQueue: taskQueue,
		logger:    logger,
		processingTimeout: processingTimeout,
		stats: stats,
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

			grouped, err := anagram.Group(spanCtx, task.Words, task.CaseSensitive)
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
