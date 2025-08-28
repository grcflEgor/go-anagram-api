package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/repository"
	"github.com/grcflEgor/go-anagram-api/pkg/anagram"
	"go.uber.org/zap"
)

const processingTimeout = 30 * time.Second

type Pool struct {
	repo repository.TaskRepository
	taskQueue chan *domain.Task
	logger *zap.Logger
	wg sync.WaitGroup
}

func NewPool(repo repository.TaskRepository, taskQueue chan *domain.Task, logger *zap.Logger) *Pool {
	return &Pool{
		repo: repo,
		taskQueue: taskQueue,
		logger: logger,
	}
}

func (p *Pool) Run(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i + 1)
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()

	workerLog := p.logger.With(zap.Int("worker_id", id))
	workerLog.Info("worker started")

	for task := range p.taskQueue {
		func(task *domain.Task) {
			taskLog := workerLog.With(zap.String("task_id", task.ID))
			taskLog.Info("processing task")

			start := time.Now()

			ctx, cancel := context.WithTimeout(context.Background(), processingTimeout)
			defer cancel()

			grouped, err := anagram.Group(ctx, task.Words)
			processingTime := time.Since(start).Milliseconds()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					taskLog.Warn("task processing timeout")
					task.Status = domain.StatusFailed
					task.Error = "task processing timeout"
				} else {
					taskLog.Error("task processing failed", zap.Error(err))
					task.Status = domain.StatusFailed
					task.Error = err.Error()
				}
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
			}

			if err := p.repo.Save(context.Background(), task); err != nil {
				taskLog.Error("failed to save completed task", zap.String("status", string(task.Status)), zap.Error(err))
			}
			taskLog.Info("finished task")
		}(task)
	}
}

func (p *Pool) Stop() {
	close(p.taskQueue)
	p.wg.Wait()
}