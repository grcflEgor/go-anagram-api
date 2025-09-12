package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Clearenv()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	require.Equal(t, ":8080", cfg.Server.Port)
	require.Equal(t, 1000, cfg.Task.QueueSize)
	require.Equal(t, 10, cfg.Worker.Count)
	require.Equal(t, 5*time.Minute, cfg.Cache.DefaultExpiration)
	require.Equal(t, 10*time.Minute, cfg.Cache.CleanupInterval)
	require.Equal(t, "anagram-api", cfg.Service.Name)
	require.Equal(t, 30*time.Second, cfg.Processing.Timeout)
	require.Equal(t, 1000, cfg.RateLimit.Requests)
	require.Equal(t, 1*time.Minute, cfg.RateLimit.Window)
	require.Equal(t, 30*time.Second, cfg.Graceful.ShutdownTimeout)
	require.Equal(t, int64(20971520), cfg.Upload.MaxFileSize)
	require.Equal(t, "application/json,application/csv,text/plain", cfg.Upload.AllowedTypes)
	require.Equal(t, 10000, cfg.Upload.BatchSize)
}

func TestLoadConfig_WithEnvOverrides(t *testing.T) {
	os.Setenv("SERVER_PORT", ":9999")
	os.Setenv("TASK_QUEUE_SIZE", "500")
	os.Setenv("NUM_WORKERS", "8")
	os.Setenv("CACHE_DEFAULT_EXPIRATION", "2m")
	os.Setenv("CACHE_CLEANUP_INTERVAL", "3m")
	os.Setenv("SERVICE_NAME", "custom-service")
	os.Setenv("PROCESSING_TIMEOUT", "45s")
	os.Setenv("RATE_LIMIT_REQUESTS", "200")
	os.Setenv("RATE_LIMIT_WINDOW", "2m")
	os.Setenv("GRACEFUL_SHUTDOWN_TIMEOUT", "10s")
	os.Setenv("UPLOAD_MAX_FILE_SIZE", "1024")
	os.Setenv("UPLOAD_ALLOWED_TYPES", "text/plain")
	os.Setenv("UPLOAD_BATCH_SIZE", "123")

	defer os.Clearenv()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	require.Equal(t, ":9999", cfg.Server.Port)
	require.Equal(t, 500, cfg.Task.QueueSize)
	require.Equal(t, 8, cfg.Worker.Count)
	require.Equal(t, 2*time.Minute, cfg.Cache.DefaultExpiration)
	require.Equal(t, 3*time.Minute, cfg.Cache.CleanupInterval)
	require.Equal(t, "custom-service", cfg.Service.Name)
	require.Equal(t, 45*time.Second, cfg.Processing.Timeout)
	require.Equal(t, 200, cfg.RateLimit.Requests)
	require.Equal(t, 2*time.Minute, cfg.RateLimit.Window)
	require.Equal(t, 10*time.Second, cfg.Graceful.ShutdownTimeout)
	require.Equal(t, int64(1024), cfg.Upload.MaxFileSize)
	require.Equal(t, "text/plain", cfg.Upload.AllowedTypes)
	require.Equal(t, 123, cfg.Upload.BatchSize)
}
