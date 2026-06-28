package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"quiz-service/model"
)

// JWT returns a middleware that validates a Bearer token and injects
// the student_id into the X-Student-ID request header.
func JWT(secret []byte, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "Invalid auth header format"}`, http.StatusUnauthorized)
			return
		}

		claims := &model.Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(_ *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-Student-ID", claims.StudentID)
		next(w, r)
	}
}
