package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   []byte
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/quiz_db?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "my_super_secret_key"
	}

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
		JWTSecret:   []byte(jwtSecret),
	}
}
