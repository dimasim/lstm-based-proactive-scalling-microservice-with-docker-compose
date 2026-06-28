package main

import (
	"log"
	"net/http"

	"quiz-service/config"
	"quiz-service/database"
	"quiz-service/handler"
	"quiz-service/middleware"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/quiz", middleware.JWT(cfg.JWTSecret, handler.Quiz(db)))
	mux.HandleFunc("/healthz", handler.Healthz)

	log.Printf("Go Quiz Service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
