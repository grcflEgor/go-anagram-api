package worker

import (
	"context"
	"sync"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/repository"
	"github.com/grcflEgor/go-anagram-api/pkg/anagram"
	"go.uber.org/zap"
)

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
		taskLog := workerLog.With(zap.String("task_id", task.ID))
		taskLog.Info("processing task")

		start := time.Now()
		grouped := anagram.Group(task.Words)
		processingTime := time.Since(start).Milliseconds()

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

		if err := p.repo.Save(context.Background(), task); err != nil {
    		taskLog.Error("failed to save completed task", zap.Error(err))
    		task.Status = domain.StatusFailed
    		task.Error = err.Error() 
		
    
    		if err2 := p.repo.Save(context.Background(), task); err2 != nil {
        		taskLog.Error("failed to save FAILED task status", zap.Error(err2))
   			}
			taskLog.Info("finished task")
		}
	}
}

func (p *Pool) Stop() {
	close(p.taskQueue)
	p.wg.Wait()
}