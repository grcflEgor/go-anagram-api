package main

import (
	"context"

	"github.com/grcflEgor/go-anagram-api/internal/config"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"github.com/grcflEgor/go-anagram-api/pkg/tracing"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	defer func() { _ = logger.AppLogger.Sync() }()

	config, err := config.LoadConfig()
	if err != nil {
		logger.AppLogger.Fatal("failed to load config", zap.Error(err))
	}

	tracerProvider, err := tracing.NewTracerProvider(logger.AppLogger, config.Service.Name)
	if err != nil {
		logger.AppLogger.Fatal("failed to initialize tracing", zap.Error(err))
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.AppLogger.Error("failed to shutdown tracing", zap.Error(err))
		}
	}()

	dependencies := NewDependencies(config)
	dependencies.Start()

	server := NewServer(config, dependencies.Handlers)
	gracefulManager := NewGracefulManager(config, server, dependencies)

	gracefulManager.RunWithGracefulShutdown()
}
