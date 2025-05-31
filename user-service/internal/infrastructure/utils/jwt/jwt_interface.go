package jwt

import (
	"context"
	"time"
)

type JWTService interface {
	GenerateJWT(userID string) (string, error)
	VerifyToken(token string) (string, error)
	BlacklistToken(token string, ttl time.Duration) error
	InvalidateToken(tokenString string) error
	ExtractTokenFromContext(ctx context.Context) (string, error)
}
