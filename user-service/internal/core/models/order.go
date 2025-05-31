package models

import "time"

type Order struct {
	ID         string      `json:"id" bson:"_id,omitempty"`
	UserID     string      `json:"user_id" bson:"user_id"`
	OrderID    string `json:"order_id" bson:"order_id"`
	Status     string      `json:"status" bson:"status"`
	TotalPrice float64     `json:"total_price" bson:"total_price"`
	CreatedAt  time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" bson:"updated_at"`
	Items      []OrderItem `json:"items" bson:"items"`
}

type OrderItem struct {
	ProductID    string  `json:"product_id" bson:"product_id"`
	Quantity     int     `json:"quantity" bson:"quantity"`
	PricePerUnit float64 `json:"price_per_unit" bson:"price_per_unit"`
}