package services

import (
	"context"
	"log"
	userpb "proto/generated/ecommerce/user"
	"time"
	"user-service/internal/core/models"
	"user-service/internal/infrastructure/cache"
	"user-service/internal/infrastructure/utils/jwt"
	logger "user-service/internal/interfaces/logger"
	"user-service/internal/usecases/services"
)

type UserGrpcServer struct {
	userpb.UnimplementedUserServiceServer
	userService *services.UserService
	tokenGen    jwt.JWTService
	logger      logger.Logger
	cache       cache.CacheService
}

func NewUserGrpcServer(
	userService *services.UserService,
	tokenGen jwt.JWTService,
	logger logger.Logger,
	cache cache.CacheService,
) *UserGrpcServer {
	return &UserGrpcServer{
		userService: userService,
		tokenGen:    tokenGen,
		logger:      logger,
		cache:       cache,
	}
}

func (s *UserGrpcServer) RegisterUser(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	user := models.User{
		Username:  req.GetUsername(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdUser, err := s.userService.RegisterUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &userpb.RegisterResponse{
		Id:        createdUser.ID,
		Username:  createdUser.Username,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserGrpcServer) LoginUser(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	user, err := s.userService.AuthenticateUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	token, err := s.tokenGen.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	return &userpb.LoginResponse{
		Token:    token,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (s *UserGrpcServer) RetrieveProfile(ctx context.Context, req *userpb.RetrieveProfileRequest) (*userpb.RetrieveProfileResponse, error) {
	userID := req.GetUserId()

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		return nil, err
	}

	resp, err := &userpb.RetrieveProfileResponse{
		Username: user.Username,
		Email:    user.Email,
	}, nil

	return resp, nil
}

func (s *UserGrpcServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {

	tokenString, err := s.tokenGen.ExtractTokenFromContext(ctx)
	if err != nil {
		s.logger.Errorf("Failed to extract token from context: %v", err)
		return nil, err
	}

	userID := req.GetUserId()

	err = s.userService.DeleteUserAndOrders(ctx, userID, tokenString)

	if err != nil {
		s.logger.Errorf("Failed to delete user and orders: %v", err)
		return nil, err
	}

	return &userpb.DeleteUserResponse{
		Message: "User and associated orders deleted successfully",
	}, nil
}
