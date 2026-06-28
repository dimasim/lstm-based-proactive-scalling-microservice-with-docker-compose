package model

import "github.com/golang-jwt/jwt/v5"

// Claims holds the JWT payload.
type Claims struct {
	StudentID string `json:"student_id"`
	jwt.RegisteredClaims
}
