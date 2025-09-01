package config

import (
	"time"
	"github.com/caarlos0/env/v10"

)

type Config struct {
	Server struct {
		Port string `env:"SERVER_PORT" envDefault:":8080"`
	}

	Task struct {
		QueueSize int `env:"TASK_QUEUE_SIZE" envDefault:"100"`
	}

	Worker struct {
		Count int `env:"NUM_WORKERS" envDefault:"4"`
	}

	Cache struct {
		DefaultExpiration time.Duration `env:"CACHE_DEFAULT_EXPIRATION" envDefault:"5m"`
		CleanupInterval   time.Duration `env:"CACHE_CLEANUP_INTERVAL" envDefault:"10m"`
	}

	Service struct {
		Name string `env:"SERVICE_NAME" envDefault:"anagram-api"`
	}

	Processing struct {
		Timeout time.Duration `env:"PROCESSING_TIMEOUT" envDefault:"30s"`
	}

	RateLimit struct {
		Requests int           `env:"RATE_LIMIT_REQUESTS" envDefault:"100"`
		Window   time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1m"`
	}

	Graceful struct {
		ShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT" envDefault:"30s"`
	}

	Upload struct {
		MaxFileSize  int64  `env:"UPLOAD_MAX_FILE_SIZE" envDefault:"20971520"`
		AllowedTypes string `env:"UPLOAD_ALLOWED_TYPES" envDefault:"application/json,application/csv,text/plain"`
		BatchSize int `env:"UPLOAD_BATCH_SIZE" envDefault:"10000"`
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, err
	}

	return config, nil
}