package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpHandlers "github.com/grcflEgor/go-anagram-api/internal/delivery/http"
	"github.com/grcflEgor/go-anagram-api/internal/domain"
	"github.com/grcflEgor/go-anagram-api/internal/repository"
	"github.com/grcflEgor/go-anagram-api/internal/usecase"
	"github.com/grcflEgor/go-anagram-api/internal/worker"
	"github.com/patrickmn/go-cache"
)

const (
	taskQueueSize = 100
	numWorkers = 4
)

func main() {
	appCache := cache.New(5*time.Minute, 10*time.Minute)


	inMemoryRepo := repository.NewInMemoryStorage()
	cachedRepo := repository.NewCachedTaskRepository(inMemoryRepo, appCache)

	taskQueue := make(chan *domain.Task, taskQueueSize)
	anagramUseCase := usecase.NewAnagramUseCase(cachedRepo, taskQueue)

	workerPool := worker.NewPool(cachedRepo, taskQueue)
	workerPool.Run(numWorkers)

	handlers := httpHandlers.NewHandlers(anagramUseCase)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.HealthCheck)
		r.Post("/anagrams/group", handlers.GroupAnagrams)
		r.Get("/anagrams/groups/{id}", handlers.GetResult)
	})

	port := ":8080"
	log.Printf("starting server on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}