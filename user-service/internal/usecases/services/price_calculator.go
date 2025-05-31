package services

import "user-service/internal/core/models"

type priceCalculator struct{}

func NewPriceCalculator() PriceCalculator {
	return &priceCalculator{}
}

func (p *priceCalculator) CalculateTotalPrice(items []models.OrderItem) float64 {
	var totalPrice float64
	for _, item := range items {
		totalPrice += float64(item.Quantity) * item.PricePerUnit
	}
	return totalPrice
}
