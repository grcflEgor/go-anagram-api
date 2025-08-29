package main

import (
	"time"
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
}

func DefaultConfig() *Config {
	config := &Config{}

	config.Server.Port = ":8080"
	config.Task.QueueSize = 100
	config.Worker.Count = 4
	config.Cache.DefaultExpiration = 5 * time.Minute
	config.Cache.CleanupInterval = 10 * time.Minute
	config.Service.Name = "anagram-api"
	config.Processing.Timeout = 30 * time.Second
	config.RateLimit.Requests = 100
	config.RateLimit.Window = 1 * time.Minute
	config.Graceful.ShutdownTimeout = 30 * time.Second

	return config
}
