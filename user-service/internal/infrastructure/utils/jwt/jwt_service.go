package jwt

import (
	"context"
	"user-service/internal/errors"
	"user-service/internal/infrastructure/cache"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type service struct {
	secretKey string
	cache cache.CacheService
}

func NewJWTService(secretKey string, cache cache.CacheService) *service {
	return &service{secretKey: secretKey, cache: cache}
}

func (s *service) GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", errors.ErrJWTGeneration
	}
	return tokenString, nil
}

func (s *service) VerifyToken(tokenString string) (string, error) {

	if ok, err := s.cache.Exists(tokenString); err == nil && ok {
		return "", errors.ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrInvalidToken
		}
		return []byte(s.secretKey), nil
	})

	if err != nil || !token.Valid {
		return "", errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.ErrInvalidToken
	}

	return userID, nil
}

func (s *service) InvalidateToken(tokenString string) error {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil || !parsedToken.Valid {
		return errors.ErrInvalidToken
	}

	claims := parsedToken.Claims.(jwt.MapClaims)
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.ErrInvalidToken
	}

	expTime := time.Unix(int64(exp), 0)
	ttl := time.Until(expTime)

	return s.BlacklistToken(tokenString, ttl)
}

func (s *service) BlacklistToken(tokenString string, ttl time.Duration) error {
	return s.cache.Set(tokenString, "blacklisted", ttl)
}

func (s *service) ExtractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not provided")
	}

	authHeaders := md["authorization"]
	if len(authHeaders) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token not found")
	}

	tokenParts := strings.SplitN(authHeaders[0], " ", 2)
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return tokenParts[1], nil
}
