package worker

import (
	"context"
	"log"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/repository"
	"github.com/grcflEgor/go-anagram-api/pkg/anagram"
)

type Pool struct {
	repo repository.TaskRepository
	taskQueue <-chan *domain.Task
}

func NewPool(repo repository.TaskRepository, taskQueue <-chan *domain.Task) *Pool {
	return &Pool{
		repo: repo,
		taskQueue: taskQueue,
	}
}

func (p *Pool) Run(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go p.worker(i + 1)
	}
}

func (p *Pool) worker(id int) {
	log.Printf("worker %d started", id)
	for task := range p.taskQueue {
		log.Printf("worker %d: processing task %s", id, task.ID)

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
			log.Printf("worker %d: failed to save task%s: %v", id, task.ID, err)
			task.Status = domain.StatusFailed
		}
		log.Printf("worker %d: finished task %s:", id, task.ID)
	}
}