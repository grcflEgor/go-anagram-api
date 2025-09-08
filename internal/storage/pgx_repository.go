package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repositories.TaskStorage = (*PostgresTaskRepo)(nil)

type PostgresTaskRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresTaskRepo(pool *pgxpool.Pool) *PostgresTaskRepo {
	return &PostgresTaskRepo{
		pool: pool,
	}
}

func (r *PostgresTaskRepo) Save(ctx context.Context, task *domain.Task) error {
	wordsJSON, err := json.Marshal(task.Words)
	if err != nil {
		return err
	}

	resultJSON, err := json.Marshal(task.Result)
	if err != nil {
		return err
	}

	if len(task.Result) == 0 {
		resultJSON = []byte("[]")
	}

	query := `
    	INSERT INTO tasks (id, status, words, result, file_path, case_sensitive, error, created_at, processing_time_ms, groups_count)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (id) DO UPDATE SET
            status = EXCLUDED.status,
            result = EXCLUDED.result,
            error = EXCLUDED.error,
            processing_time_ms = EXCLUDED.processing_time_ms,
            groups_count = EXCLUDED.groups_count;
    `

	_, err = r.pool.Exec(ctx, query,
		task.ID,
		task.Status,
		wordsJSON,
		resultJSON,
		task.FilePath,
		task.CaseSensitive,
		task.Error,
		task.CreatedAt,
		task.ProcessingTimeMS,
		task.GroupsCount,
	)
	return err
}

func (r *PostgresTaskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, status, words, result, file_path, case_sensitive, error, created_at, processing_time_ms, groups_count
		FROM tasks WHERE id = $1;
	`

	task := &domain.Task{}
	var wordsJSON, resultJSON []byte

	 err := r.pool.QueryRow(ctx, query, id).Scan(
        &task.ID,
        &task.Status,
        &wordsJSON,
        &resultJSON,
        &task.FilePath,
        &task.CaseSensitive,
        &task.Error,
        &task.CreatedAt,
        &task.ProcessingTimeMS,
        &task.GroupsCount,
    )

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("not found with id %s", task.ID)
		}
	}

	if err := json.Unmarshal(wordsJSON, &task.Words); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resultJSON, &task.Result); err != nil {
		return nil, err
	}

	return task, nil
}