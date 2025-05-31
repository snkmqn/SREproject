package middleware

import (
	"context"
	"user-service/internal/infrastructure/utils/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func JWTInterceptor(jwtService jwt.JWTService) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		switch info.FullMethod {
		case "/ecommerce/.InventoryService/CreateProduct",
			"/ecommerce/.InventoryService/UpdateProduct",
			"/ecommerce/.InventoryService/DeleteProduct",
			"/user.UserService/RetrieveProfile",
			"/ecommerce/.order.OrderService/CreateOrder",
			"/ecommerce/.order.OrderService/GetOrderByID",
			"/ecommerce/.order.OrderService/UpdateOrder",
			"/ecommerce/.order.OrderService/GetOrderByUserID":

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
			}

			authHeader := md["authorization"]
			if len(authHeader) == 0 {
				return nil, status.Errorf(codes.Unauthenticated, "authorization token is required")
			}

			token := authHeader[0]

			if len(token) < 7 || token[:7] != "Bearer " {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token format")
			}
			token = token[7:]

			if token == "" {
				return nil, status.Errorf(codes.Unauthenticated, "missing token")
			}

			_, err := jwtService.VerifyToken(token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
			}
		}
		return handler(ctx, req)
	}
}
