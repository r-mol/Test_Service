package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"user_id"`
	IP     string `json:"ip"`
	jwt.RegisteredClaims
}

type Payload struct {
	UserID string `json:"user_id"`
	IP     string `json:"ip"`
}

type Token struct {
	ID        int
	UserID    uuid.UUID
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
