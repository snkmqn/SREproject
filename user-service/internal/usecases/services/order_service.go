package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	inventorypb "proto/generated/ecommerce/inventory"
	orderpb "proto/generated/ecommerce/order"
	"time"
	"user-service/internal/core/models"
	"user-service/internal/infrastructure/cache"
	"user-service/internal/infrastructure/utils/uuid"
	logger "user-service/internal/interfaces/logger"
	"user-service/internal/interfaces/repositories"
)

type OrderService struct {
	orderRepo       repositories.OrderRepository
	productService  *ProductService
	priceCalculator PriceCalculator
	uuidGenerator   *uuid.Service
	cache           cache.CacheService
	logger          logger.Logger
}

func NewOrderService(orderRepo repositories.OrderRepository, priceCalculator PriceCalculator, uuidGenerator *uuid.Service, ProductService *ProductService, cache cache.CacheService, logger logger.Logger) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		productService:  ProductService,
		priceCalculator: priceCalculator,
		uuidGenerator:   uuidGenerator,
		cache:           cache,
		logger:          logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []models.OrderItem, status string) (*models.Order, error) {
	totalPrice := s.priceCalculator.CalculateTotalPrice(items)

	order := &models.Order{
		ID:         s.uuidGenerator.GenerateUUID(),
		OrderID:    s.uuidGenerator.GenerateUUID(),
		UserID:     userID,
		Status:     status,
		Items:      items,
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) CreateOrderFromProto(ctx context.Context, req *orderpb.CreateOrderRequest) (*models.Order, error) {
	var items []models.OrderItem

	for _, item := range req.GetItems() {
		if item.GetQuantity() <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "quantity for %s must be > 0", item.GetProductId())
		}
		if item.GetPricePerUnit() <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "price per unit for %s must be > 0", item.GetProductId())
		}

		stock, err := s.CheckProductStock(ctx, item.GetProductId(), int(item.GetQuantity()))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "stock check failed: %v", err)
		}

		if !stock.InStock {
			return nil, status.Errorf(codes.FailedPrecondition, "product %s is out of stock (%d available)", item.GetProductId(), stock.AvailableStock)
		}

		items = append(items, models.OrderItem{
			ProductID:    item.GetProductId(),
			Quantity:     int(item.GetQuantity()),
			PricePerUnit: item.GetPricePerUnit(),
		})
	}

	return s.CreateOrder(ctx, req.GetUserId(), items, req.GetStatus())
}

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	cacheKey := fmt.Sprintf("order:%s", id)

	cachedData, err := s.cache.Get(cacheKey)
	if err != nil {
		s.logger.Errorf("Redis error: %v", err)
	}

	if err == nil && cachedData != "" {
		var cachedOrder models.Order
		err := json.Unmarshal([]byte(cachedData), &cachedOrder)
		if err == nil {
			s.logger.Infof("Cache hit for order %s", id)
			return &cachedOrder, nil
		}
		s.logger.Errorf("Failed to unmarshal cached order data: %v", err)
	}

	s.logger.Infof("Cache miss for order %s", id)

	order, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(order)
	if err == nil {
		err = s.cache.Set(cacheKey, string(jsonData), 10*time.Minute)
		if err != nil {
			s.logger.Errorf("Failed to set order to cache: %v", err)
		}
	}

	return order, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, id string, status string) error {
	if status != "pending" && status != "completed" && status != "cancelled" {
		return errors.New("invalid status")
	}

	cacheKey := fmt.Sprintf("order:%s", id)

	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Infof("Failed to invalidate cache for order %s: %v", id, err)
	}

	err := s.orderRepo.UpdateOrder(ctx, id, status)
	return err
}

func (s *OrderService) GetOrdersByUserID(ctx context.Context, userID string) ([]*models.Order, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) CheckProductStock(ctx context.Context, productID string, quantity int) (*inventorypb.CheckStockResponse, error) {
	inStock, availableStock, err := s.productService.CheckStock(ctx, productID, int32(quantity))
	if err != nil {
		return nil, err
	}

	return &inventorypb.CheckStockResponse{
		InStock:        inStock,
		AvailableStock: availableStock,
	}, nil
}
