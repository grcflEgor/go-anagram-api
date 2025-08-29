package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	httpHandlers "github.com/grcflEgor/go-anagram-api/internal/controller/http/v1"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"go.uber.org/zap"
)

type Server struct {
	config     *Config
	handlers   *httpHandlers.Handlers
	httpServer *http.Server
}

func NewServer(config *Config, handlers *httpHandlers.Handlers) *Server {
	router := setupRouter(config, handlers)

	server := &http.Server{
		Addr:         config.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		config:     config,
		handlers:   handlers,
		httpServer: server,
	}
}

func setupRouter(config *Config, handlers *httpHandlers.Handlers) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(httpHandlers.LoggerMiddleware)
	router.Use(httpHandlers.OTelMiddleware)
	router.Use(httprate.Limit(
		config.RateLimit.Requests,
		config.RateLimit.Window,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.HealthCheck)
		r.Post("/anagrams/group", handlers.GroupAnagrams)
		r.Get("/anagrams/groups/{id}", handlers.GetResult)
	})

	return router
}

func (s *Server) Start() error {
	logger.AppLogger.Info("starting HTTP server",
		zap.String("port", s.config.Server.Port))

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	logger.AppLogger.Info("shutting down HTTP server")
	return s.httpServer.Close()
}

func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}
