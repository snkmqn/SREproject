package repositories

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"user-service/internal/core/models"
	"user-service/internal/interfaces/repositories"
)

type orderRepositoryMongo struct {
	collection *mongo.Collection
}

func NewOrderRepositoryMongo(db *mongo.Database) repositories.OrderRepository {
	return &orderRepositoryMongo{
		collection: db.Collection("orders"),
	}
}

func (r *orderRepositoryMongo) CreateOrder(ctx context.Context, order *models.Order) error {
	_, err := r.collection.InsertOne(ctx, order)
	fmt.Println("Saving order to DB:", order)
	return err
}

func (r *orderRepositoryMongo) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	fmt.Println("Querying order with ID:", id)
	err := r.collection.FindOne(ctx, bson.M{"order_id": id}).Decode(&order)
	if err != nil {
		return nil, err
	}
	fmt.Println("Found order:", order)
	return &order, nil
}

func (r *orderRepositoryMongo) UpdateOrder(ctx context.Context, id string, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}

func (r *orderRepositoryMongo) GetOrdersByUserID(ctx context.Context, userID string) ([]*models.Order, error) {
	var orders []*models.Order
	fmt.Println("Querying order with ID:", userID)

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}

		orders = append(orders, &order)

	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepositoryMongo) DeleteOrdersByUserID(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
