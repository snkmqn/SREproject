package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
	"user-service/internal/core/models"
	"user-service/internal/errors"
	"user-service/internal/infrastructure/cache"
	logger "user-service/internal/interfaces/logger"
	"user-service/internal/interfaces/repositories"
	"user-service/internal/usecases/validators"
)

type ProductService struct {
	productRepo repositories.ProductRepository
	logger      logger.Logger
	cache       cache.CacheService
}

func NewProductService(productRepo repositories.ProductRepository, logger logger.Logger, cache cache.CacheService) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		logger:      logger,
		cache:       cache,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product models.Product) (models.Product, error) {
	if err := validators.ValidateProductForCreation(product); err != nil {
		s.logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return models.Product{}, err
	}
	if product.ID == "" {
		product.ID = uuid.NewString()
	}

	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	createdProduct, err := s.productRepo.CreateProduct(ctx, product)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error creating product: %v", err))
		return models.Product{}, err
	}

	err = s.cache.InvalidateKeysByPrefix("products:")
	if err != nil {
		s.logger.Errorf("Failed to invalidate product cache: %v", err)
	}
	s.logger.Info(fmt.Sprintf("Product %s created successfully", createdProduct.Name))
	return createdProduct, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, product models.Product) (models.Product, error) {
	if err := validators.ValidateProductForUpdate(product); err != nil {
		s.logger.Error(fmt.Sprintf("Product update validation failed: %v", err))
		return models.Product{}, err
	}

	product.UpdatedAt = time.Now()

	updatedProduct, err := s.productRepo.UpdateProduct(ctx, id, product)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to update product ID %s: %v", id, err))
		return models.Product{}, err
	}

	err = s.cache.InvalidateKeysByPrefix("products:")
	if err != nil {
		s.logger.Errorf("Failed to invalidate product cache: %v", err)
	}

	updatedProduct.ID = id
	s.logger.Info(fmt.Sprintf("Product ID %s updated successfully", id))
	return updatedProduct, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (models.Product, error) {
	return s.productRepo.GetProductByID(ctx, id)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {

	err := s.cache.InvalidateKeysByPrefix("products:")
	if err != nil {
		s.logger.Errorf("Failed to invalidate product cache: %v", err)
	}
	return s.productRepo.DeleteProduct(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]models.Product, error) {
	filterBytes, _ := json.Marshal(filter)

	cacheKey := fmt.Sprintf("products:filter=%s:skip=%d:limit=%d", string(filterBytes), skip, limit)

	cachedData, err := s.cache.Get(cacheKey)

	if err == nil && cachedData != "" {
		var cachedProducts []models.Product
		err = json.Unmarshal([]byte(cachedData), &cachedProducts)

		if err == nil {
			s.logger.Infof("Cache hit for product list")
			return cachedProducts, nil
		}

		s.logger.Errorf("Failed to unmarshal cached product list: %v", err)
	}

	s.logger.Info("Cache miss for products")

	products, err := s.productRepo.ListProducts(ctx, filter, skip, limit)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(products)
	if err == nil {
		err = s.cache.Set(cacheKey, string(jsonData), 10*time.Minute)
		if err != nil {
			s.logger.Errorf("Failed to set product list to cache: %v", err)
		}
	}
	return products, nil
}

func (s *ProductService) CheckStock(ctx context.Context, productID string, quantity int32) (bool, int32, error) {
	if productID == "" {
		s.logger.Error("CheckStock: product ID is empty")
		return false, 0, fmt.Errorf("product ID is required")
	}

	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("CheckStock: failed to get product by ID %s: %v", productID, err))
		return false, 0, fmt.Errorf("product not found: %w", err)
	}

	inStock := product.Stock >= int(quantity)

	s.logger.Info(fmt.Sprintf(
		"CheckStock: product_id=%s, requested=%d, available=%d, in_stock=%v",
		productID, quantity, product.Stock, inStock,
	))

	return inStock, int32(product.Stock), nil
}

func (s *ProductService) DecreaseStock(ctx context.Context, productID string, quantity int32) (*models.Product, error) {

	log.Printf("DecreaseStock called with ProductID=%s, Quantity=%d", productID, quantity)

	if productID == "" || quantity <= 0 {
		s.logger.Error("DecreaseStock: invalid request")
		return nil, fmt.Errorf("invalid request")
	}

	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", errors.ErrProductNotFound)
	}

	if product.Stock < int(quantity) {
		s.logger.Error(fmt.Sprintf("DecreaseStock: insufficient stock for product %s", product.ID))
		return nil, errors.ErrInsufficientStock
	}

	product.Stock -= int(quantity)

	updatedProduct, err := s.productRepo.UpdateProduct(ctx, product.ID, product)
	if err != nil {
		return nil, err
	}
	return &updatedProduct, nil
}
