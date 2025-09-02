package main

import (
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/docs/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("docs", r.URL.Path[6:]))
	})

	http.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.yaml")
	})

	http.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	log.Println("Документация доступна по адресу: http://localhost:8081")
	log.Println("Swagger UI: http://localhost:8081")
	log.Println("Swagger YAML: http://localhost:8081/swagger.yaml")
	log.Println("Swagger JSON: http://localhost:8081/swagger.json")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
