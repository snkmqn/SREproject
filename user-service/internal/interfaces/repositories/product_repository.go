package repositories

import (
	"context"
	"user-service/internal/core/models"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product models.Product) (models.Product, error)
	GetProductByID(ctx context.Context, id string) (models.Product, error)
	UpdateProduct(ctx context.Context, id string, product models.Product) (models.Product, error)
	DeleteProduct(ctx context.Context, id string) error
	ListProducts(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]models.Product, error)
	DecreaseStock(ctx context.Context, productID string, quantity int) (models.Product, error)
}
