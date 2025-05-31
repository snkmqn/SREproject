package repositories

import (
	"context"
	"user-service/internal/core/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	AuthenticateUser(ctx context.Context, email, password string) (models.User, error)
	GetUserByID(ctx context.Context, userID string) (models.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

