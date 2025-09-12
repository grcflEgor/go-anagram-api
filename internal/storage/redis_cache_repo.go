package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/domain/repositories"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var _ repositories.CacheTaskStorage = (*RedisCachedStorage)(nil)

type RedisCachedStorage struct {
	client *redis.Client
	ttl time.Duration
}

func NewRedisCachedStorage(client *redis.Client, ttl time.Duration) *RedisCachedStorage {
	return &RedisCachedStorage{
		client: client,
		ttl: ttl,
	}
}

func (c *RedisCachedStorage) key(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func (c *RedisCachedStorage) Save(ctx context.Context, task *domain.Task) error {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("repository")
	ctx, span := tr.Start(ctx, "RedisCachedStorage.Save")
	defer span.End()

	data, err := json.Marshal(task)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("serialize err %w", err)
	}

	if err := c.client.Set(ctx, c.key(task.ID), data, c.ttl).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("save into redis err: %w", err)
	}

	l.Info("cache saved")
	return nil
}

func (c *RedisCachedStorage) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("repository")
	ctx, span := tr.Start(ctx, "RedisCachedStorage.GetById")
	defer span.End()

	result, err := c.client.Get(ctx, c.key(id)).Result()
	if err == redis.Nil {
		l.Info("cache MISS for task id", zap.String("id", id))
		return nil, nil
	} else {
		if err != nil {
			span.RecordError(err)
			l.Error("err for get data from redis", zap.Error(err))
			return nil, err
		}
	}

	var task domain.Task
	if err := json.Unmarshal([]byte(result), &task); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("deserialize err %w", err)
	}

	return &task, nil
}

func (c *RedisCachedStorage) Delete(ctx context.Context, id string) error {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("repository")
	ctx, span := tr.Start(ctx, "RedisCachedStorage.Delete")
	defer span.End()

	if err := c.client.Del(ctx, c.key(id)).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("del err from redis %w", err)
	}

	l.Info("delete succsed")
	return nil
}

func (c *RedisCachedStorage) DeleteAll(ctx context.Context) error {
	l := logger.FromContext(ctx)

	tr := otel.Tracer("repository")
	ctx, span := tr.Start(ctx, "RedisCachedStorage.DeleteALL")
	defer span.End()

	if err := c.client.FlushAllAsync(ctx).Err(); err != nil {
		span.RecordError(err)
		l.Error("flush all cache err", zap.Error(err))
		return err
	}
	return nil
}