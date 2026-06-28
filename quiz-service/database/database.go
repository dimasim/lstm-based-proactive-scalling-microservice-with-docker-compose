package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Connect opens a Postgres connection with a simple retry loop.
func Connect(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			return db, nil
		}
		log.Printf("Waiting for postgres... attempt %d/5: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

// Migrate runs any required DDL statements to bootstrap the schema.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS quiz_attempts (
		id         SERIAL PRIMARY KEY,
		student_id VARCHAR(50)  NOT NULL,
		timestamp  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}
