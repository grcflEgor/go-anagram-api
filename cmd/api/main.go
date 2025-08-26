package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpHandlers "github.com/grcflEgor/go-anagram-api/internal/delivery/http"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", httpHandlers.HealthCheckHandler)
		r.Post("/anagrams/group", httpHandlers.GroupAnagramsHandler)
		r.Get("/anagrams/groups/{id}", httpHandlers.GetResultHandler)
	})

	port := ":8080"
	log.Printf("starting server on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}