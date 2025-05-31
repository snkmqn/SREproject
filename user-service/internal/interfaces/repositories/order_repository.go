package repositories

import (
	"context"
	"user-service/internal/core/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByID(ctx context.Context, id string) (*models.Order, error)
	UpdateOrder(ctx context.Context, id string, status string) error
	GetOrdersByUserID (ctx context.Context, userID string) ([]*models.Order, error)
	DeleteOrdersByUserID (ctx context.Context, userID string) error
}
