package services

import (
	"context"
	"user-service/internal/core/models"
	"log"
)

type StockDecrementer struct {
	productService ProductService
}

func NewStockDecrementer(p ProductService) *StockDecrementer {
	return &StockDecrementer{productService: p}
}

func (s *StockDecrementer) HandleOrder(ctx context.Context, order models.Order) error {

	for _, item := range order.Items {
		updatedProduct, err := s.productService.DecreaseStock(ctx, item.ProductID, int32(item.Quantity))

		if err != nil {
			log.Printf("Failed to decrease stock for product %s: %v", item.ProductID, err)
			return err
		}

		log.Printf("Stock decreased successfully for product %s, new stock: %d", item.ProductID, updatedProduct.Stock)
	}
	return nil
}
