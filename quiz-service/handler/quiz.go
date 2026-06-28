package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

// Quiz handles POST/GET /quiz requests.
// It logs the attempt to the database and returns a simple JSON response.
func Quiz(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		studentID := r.Header.Get("X-Student-ID")

		if _, err := db.Exec("INSERT INTO quiz_attempts (student_id) VALUES ($1)", studentID); err != nil {
			log.Printf("DB error logging attempt: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"success","message":"Quiz loaded","student_id":"%s"}`, studentID)
	}
}

// Healthz is a simple liveness probe.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
