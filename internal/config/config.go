// Package config provides configuration management for the anagram API service.
// It uses environment variables to configure various aspects of the application.
package config

import (
	"time"
	"github.com/caarlos0/env/v10"

)

// Config holds all configuration settings for the anagram API service.
// All fields are populated from environment variables using struct tags.
type Config struct {
	// Server configuration for HTTP server settings
	Server struct {
		Port string `env:"SERVER_PORT" envDefault:":8080"` // HTTP server port
	}

	// Task configuration for task processing
	Task struct {
		QueueSize int `env:"TASK_QUEUE_SIZE" envDefault:"1000"` // Maximum number of tasks in queue
	}

	// Worker configuration for worker pool
	Worker struct {
		Count int `env:"NUM_WORKERS" envDefault:"10"` // Number of worker goroutines
	}

	// Cache configuration for in-memory caching
	Cache struct {
		DefaultExpiration time.Duration `env:"CACHE_DEFAULT_EXPIRATION" envDefault:"5m"` // Default cache entry expiration
		CleanupInterval   time.Duration `env:"CACHE_CLEANUP_INTERVAL" envDefault:"10m"` // Cache cleanup interval
	}

	// Service configuration for service identification
	Service struct {
		Name string `env:"SERVICE_NAME" envDefault:"anagram-api"` // Service name for logging and tracing
	}

	// Processing configuration for task processing timeouts
	Processing struct {
		Timeout time.Duration `env:"PROCESSING_TIMEOUT" envDefault:"30s"` // Maximum time for processing a single task
	}

	// RateLimit configuration for API rate limiting
	RateLimit struct {
		Requests int           `env:"RATE_LIMIT_REQUESTS" envDefault:"1000"` // Maximum requests per window
		Window   time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1m"`     // Rate limit time window
	}

	// Graceful shutdown configuration
	Graceful struct {
		ShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT" envDefault:"30s"` // Maximum time to wait for graceful shutdown
	}

	// Upload configuration for file upload settings
	Upload struct {
		MaxFileSize  int64  `env:"UPLOAD_MAX_FILE_SIZE" envDefault:"20971520"`    // Maximum file size in bytes (20MB)
		AllowedTypes string `env:"UPLOAD_ALLOWED_TYPES" envDefault:"application/json,application/csv,text/plain"` // Comma-separated list of allowed MIME types
		BatchSize int `env:"UPLOAD_BATCH_SIZE" envDefault:"10000"`          // Batch size for processing large files
	}

	// DBconfig configuration for PostgreSQL database connection pool
	DBconfig struct {
		MaxConns int `env:"DB_MAX_CONNS" envDefault:"25"` // Maximum number of database connections
    	MinConns int `env:"DB_MIN_CONNS" envDefault:"5"`  // Minimum number of database connections
	}

	// Redis configuration for Redis cache connection
	Redis struct {
    	URL      string `env:"REDIS_URL" envDefault:"redis://localhost:6379/0"` // Redis connection URL
        Password string `env:"REDIS_PASSWORD"`                                      // Redis password (optional)
        PoolSize int    `env:"REDIS_POOL_SIZE" envDefault:"50"`                     // Redis connection pool size
    }
}

// LoadConfig loads configuration from environment variables.
// Returns a populated Config struct or an error if parsing fails.
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Parse environment variables into the config struct
	if err := env.Parse(config); err != nil {
		return nil, err
	}

	return config, nil
}