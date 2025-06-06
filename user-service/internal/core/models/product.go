package models

import (
	"time"
)

type Product struct {
	ID          string  `json:"id" bson:"_id,omitempty"`
	Name        string     `json:"name" bson:"name"`
	Description string     `json:"description" bson:"description"`
	Price       float64    `json:"price" bson:"price"`
	Stock       int        `json:"stock" bson:"stock"`
	CategoryID  string  `json:"category_id" bson:"category_id"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
}
