package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/grcflEgor/go-anagram-api/internal/config"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"go.uber.org/zap"
)

type GracefulManager struct {
	config *config.Config
	server *Server
	deps   *Dependencies
}

func NewGracefulManager(config *config.Config, server *Server, deps *Dependencies) *GracefulManager {
	return &GracefulManager{
		config: config,
		server: server,
		deps:   deps,
	}
}

func (gm *GracefulManager) Start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.AppLogger.Info("received shutdown signal, starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), gm.config.Graceful.ShutdownTimeout)
	defer cancel()

	if err := gm.server.GetHTTPServer().Shutdown(ctx); err != nil {
		logger.AppLogger.Error("HTTP server forced to shutdown", zap.Error(err))
	} else {
		logger.AppLogger.Info("HTTP server gracefully stopped")
	}

	gm.deps.Stop()

	logger.AppLogger.Info("app gracefully shutdown completed")
}

func (gm *GracefulManager) RunWithGracefulShutdown() {
	go func() {
		if err := gm.server.Start(); err != nil {
			logger.AppLogger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	gm.Start()
}
