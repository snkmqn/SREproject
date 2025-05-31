package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"
	"user-service/internal/core/models"
	"user-service/internal/infrastructure/cache"
	jwt "user-service/internal/infrastructure/utils/jwt"
	"user-service/internal/infrastructure/utils/security"
	"user-service/internal/infrastructure/utils/uuid"
	logger "user-service/internal/interfaces/logger"
	"user-service/internal/interfaces/repositories"
	"user-service/internal/interfaces/services"
	"user-service/internal/usecases/validators"
)

type UserService struct {
	userRepo      repositories.UserRepository
	userValidator validators.UserValidator
	passwordHash  security.PasswordHash
	jwtService    jwt.JWTService
	uuidGenerator uuid.Generator
	client        *mongo.Client
	orderRepo     repositories.OrderRepository
	cache         cache.CacheService
	logger        logger.Logger
	email         services.EmailService
}

func NewUserService(userRepo repositories.UserRepository, userValidator validators.UserValidator,
	hash security.PasswordHash, jwtService jwt.JWTService, uuidGenerator uuid.Generator, client *mongo.Client, orderRepo repositories.OrderRepository,
	cache cache.CacheService, logger logger.Logger, email services.EmailService) *UserService {
	return &UserService{
		userRepo:      userRepo,
		userValidator: userValidator,
		passwordHash:  hash,
		jwtService:    jwtService,
		uuidGenerator: uuidGenerator,
		client:        client,
		orderRepo:     orderRepo,
		cache:         cache,
		logger:        logger,
		email:         email,
	}
}

func (u *UserService) RegisterUser(ctx context.Context, user models.User) (models.User, error) {

	err := u.userValidator.Validate(user)
	if err != nil {
		return models.User{}, err
	}

	existingUser, err := u.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return models.User{}, err
	}
	if existingUser.ID != "" {
		return models.User{}, errors.New("user with this email already exists")
	}

	existingUsername, err := u.userRepo.GetUserByUsername(ctx, user.Username)
	if err != nil {
		return models.User{}, err
	}
	if existingUsername.ID != "" {
		return models.User{}, errors.New("user with this username already exists")
	}

	hashedPassword, err := u.passwordHash.HashPassword(user.Password)
	if err != nil {
		return models.User{}, err
	}
	user.Password = hashedPassword

	user.ID = u.uuidGenerator.GenerateUUID()
	createdUser, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return models.User{}, err
	}

	_ = u.email.SendWelcomeEmail(user.Email)

	return createdUser, nil
}

func (u *UserService) AuthenticateUser(ctx context.Context, email, password string) (models.User, error) {

	user, err := u.userRepo.AuthenticateUser(ctx, email, password)
	if err != nil {
		return models.User{}, err
	}

	if !u.passwordHash.CheckPasswordHash(password, user.Password) {
		return models.User{}, errors.New("incorrect password")
	}
	return user, nil
}

func (u *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	cacheKey := fmt.Sprintf("user_profile:%s", userID)

	cachedData, err := u.cache.Get(cacheKey)
	if err == nil && cachedData != "" {
		var cachedUser models.User
		if err := json.Unmarshal([]byte(cachedData), &cachedUser); err == nil {
			u.logger.Infof("Cache hit for user: %s", userID)
			return &cachedUser, nil
		}
		u.logger.Error("Failed to unmarshal cached data")
	}

	u.logger.Infof("Cache miss for user %s", userID)

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return &user, err
	}

	if user.ID == "" {
		return &user, errors.New("user not found")
	}

	data, err := json.Marshal(user)
	if err == nil {
		_ = u.cache.Set(cacheKey, string(data), 10*time.Minute)
	}

	return &user, nil
}

func (s *UserService) DeleteUserAndOrders(ctx context.Context, usedID string, tokenString string) error {
	session, err := s.client.StartSession()

	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	if strings.TrimSpace(usedID) == "" {
		return status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.GetUserByID(ctx, usedID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to check user existence: %v", err)
	}

	if user == nil {
		return status.Errorf(codes.NotFound, "user with id %s not found", usedID)
	}

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err := s.orderRepo.DeleteOrdersByUserID(sessCtx, usedID); err != nil {
			return nil, err
		}

		if err := s.userRepo.DeleteUser(sessCtx, usedID); err != nil {
			return nil, err
		}

		if err := s.jwtService.InvalidateToken(tokenString); err != nil {
			return nil, err
		}
		return nil, nil
	}

	_, err = session.WithTransaction(ctx, callback)

	if err != nil {
		return err
	}

	return nil
}
