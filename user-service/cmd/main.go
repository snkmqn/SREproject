package main

import (
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	userpb "proto/generated/ecommerce/user"
	"user-service/internal/config"
	"user-service/internal/delivery/grpc/middleware"
	"user-service/internal/infrastructure/cache"
	"user-service/internal/infrastructure/database"
	"user-service/internal/infrastructure/email"
	"user-service/internal/infrastructure/logger"
	"user-service/internal/infrastructure/repositories"
	"user-service/internal/infrastructure/utils/jwt"
	"user-service/internal/infrastructure/utils/security"
	"user-service/internal/infrastructure/utils/uuid"
	grpc2 "user-service/internal/interfaces/grpc"
	repositories2 "user-service/internal/interfaces/repositories"
	"user-service/internal/usecases/services"
	"user-service/internal/usecases/validators"
)

func initRepositories() (repositories2.UserRepository, repositories2.OrderRepository, *mongo.Client, error) {

	client, err := database.ConnectMongoClient()

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to mongo client: %v", err)
	}

	userDB := client.Database("users")
	orderDB := client.Database("orders")

	userRepo := repositories.NewUserRepositoryMongo(userDB, security.NewBcryptHash())
	orderRepo := repositories.NewOrderRepositoryMongo(orderDB)

	return userRepo, orderRepo, client, nil
}

func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Prometheus metrics available on :8081/metrics")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Failed to start metrics server: %v", err)
	}
}

func main() {
	userRepo, orderRepo, client, err := initRepositories()
	if err != nil {
		log.Fatal(err)
	}

	secretKey := config.GetEnv("JWT_SECRET_KEY", "")
	stdLogger := &logger.StdLogger{}

	userValidator := validators.NewUserValidator()
	passwordHash := security.NewBcryptHash()
	uuidGen := uuid.NewUUIDService()

	redisAddr := config.GetEnv("REDIS_ADDR", "")
	redisPassword := config.GetEnv("REDIS_PASSWORD", "")
	redisDB := 0
	redisClient := cache.NewRedisCache(redisAddr, redisPassword, redisDB)

	var jwtService jwt.JWTService = jwt.NewJWTService(secretKey, redisClient)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			middleware.JWTInterceptor(jwtService),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
		),
	)

	emailService := email.NewSMTPEmailService()

	userService := services.NewUserService(userRepo, userValidator, passwordHash, jwtService, uuidGen, client, orderRepo, redisClient, stdLogger, emailService)
	userServer := grpc2.NewUserGrpcServer(userService, jwtService, stdLogger, redisClient)
	userpb.RegisterUserServiceServer(grpcServer, userServer)

	go startMetricsServer()

	grpc_prometheus.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}
	log.Println("User Service is running on port :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
