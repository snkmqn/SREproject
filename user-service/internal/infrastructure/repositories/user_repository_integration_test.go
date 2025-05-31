//go:build integration
// +build integration

package repositories

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
	"user-service/internal/core/models"
	"user-service/internal/infrastructure/cache"
	email2 "user-service/internal/infrastructure/email"
	"user-service/internal/infrastructure/utils/jwt"
	"user-service/internal/infrastructure/utils/security"
	"user-service/internal/infrastructure/utils/uuid"
	"user-service/internal/interfaces/logger"
	"user-service/internal/interfaces/repositories"
	"user-service/internal/usecases/services"
	"user-service/internal/usecases/validators"
)

func TestCreateUser_Integration(t *testing.T) {
	ctx := context.Background()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		t.Fatal("MONGODB_URI environment variable is not set")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	assert.NoError(t, err)
	defer client.Disconnect(ctx)

	db := client.Database("test")
	collection := db.Collection("users")
	defer collection.Drop(ctx)

	passwordHash := security.NewBcryptHash()
	userRepo := NewUserRepositoryMongo(db, passwordHash)
	validator := validators.NewUserValidator()
	uuidGen := uuid.NewUUIDService()
	email := email2.NewSMTPEmailService()

	var (
		jwtService jwt.JWTService               = nil
		orderRepo  repositories.OrderRepository = nil
		cache      cache.CacheService           = nil
		logger     logger.Logger                = nil
	)

	userService := services.NewUserService(userRepo, validator, passwordHash, jwtService, uuidGen, client, orderRepo, cache, logger, email)

	user := models.User{
		Username: "arsen",
		Email:    "test@xample.com",
		Password: "Password123",
	}

	start := time.Now()
	createdUser, err := userService.RegisterUser(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, createdUser.Username)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.NotEmpty(t, createdUser.ID)
	assert.WithinDuration(t, start, createdUser.CreatedAt, time.Second*5)
	assert.NotEqual(t, user.Password, createdUser.Password)

	match := passwordHash.CheckPasswordHash(user.Password, createdUser.Password)
	assert.True(t, match, "Password hash does not match original password")

	var foundUser models.User

	err = collection.FindOne(ctx, map[string]interface{}{"email": user.Email}).Decode(&foundUser)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, foundUser.ID)

	_, err = userService.RegisterUser(ctx, user)
	assert.Error(t, err)

	invalidUser := models.User{
		Username: "salem",
		Email:    "invalid@example.com",
		Password: "pass",
	}
	_, err = userService.RegisterUser(ctx, invalidUser)
	assert.Error(t, err)
}
