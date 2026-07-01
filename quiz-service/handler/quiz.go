package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Question represents a quiz question structure
type Question struct {
	ID      int      `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
}

// SubmitRequest represents the incoming JSON payload for submitting an answer
type SubmitRequest struct {
	QuestionID     int    `json:"question_id"`
	SelectedOption string `json:"selected_option"`
}

// SubmitResponse represents the JSON output returning correctness and the next question
type SubmitResponse struct {
	Status       string    `json:"status"`
	Correct      bool      `json:"correct"`
	NextQuestion *Question `json:"next_question,omitempty"`
}

// In-memory quiz questions
var quizQuestions = []Question{
	{
		ID:      1,
		Text:    "Apa kepanjangan dari CPU?",
		Options: []string{"Central Processing Unit", "Computer Personal Unit", "Control Process Utility", "Central Processor Union"},
	},
	{
		ID:      2,
		Text:    "Manakah yang merupakan bahasa pemrograman compiled language?",
		Options: []string{"Go", "Python", "Javascript", "PHP"},
	},
	{
		ID:      3,
		Text:    "Manakah database engine yang bersifat Relational (RDBMS)?",
		Options: []string{"PostgreSQL", "MongoDB", "Redis", "Cassandra"},
	},
}

// In-memory correct answers mapping (Question ID -> Correct Option Text)
var correctAnswers = map[int]string{
	1: "Central Processing Unit",
	2: "Go",
	3: "PostgreSQL",
}

// Quiz handles POST /quiz requests.
// It parses the submitted answer, validates it in-memory, logs the attempt to DB, and returns the result with the next question.
func Quiz(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, `{"status":"error","message":"Only POST method is allowed"}`)
			return
		}

		studentID := r.Header.Get("X-Student-ID")

		var req SubmitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"status":"error","message":"Invalid JSON payload"}`)
			return
		}

		// 1. Validate correctness in-memory
		correctAnswer, exists := correctAnswers[req.QuestionID]
		isCorrect := false
		if exists && req.SelectedOption == correctAnswer {
			isCorrect = true
		}

		// 2. Log the attempt to the database (1 write operation)
		if _, err := db.Exec("INSERT INTO quiz_attempts (student_id) VALUES ($1)", studentID); err != nil {
			log.Printf("DB error logging attempt: %v", err)
		}

		// 3. Find the next question
		var nextQ *Question
		for i, q := range quizQuestions {
			if q.ID == req.QuestionID {
				if i+1 < len(quizQuestions) {
					nextQ = &quizQuestions[i+1]
				}
				break
			}
		}

		// If the question ID is not found or it's the first initialization (e.g. question_id = 0), send the first question
		if req.QuestionID == 0 && len(quizQuestions) > 0 {
			nextQ = &quizQuestions[0]
		}

		resp := SubmitResponse{
			Status:       "success",
			Correct:      isCorrect,
			NextQuestion: nextQ,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// Healthz is a simple liveness probe.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

