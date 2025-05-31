package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
	"user-service/internal/config"
)

func ConnectMongoDB(dbName string) (*mongo.Database, error) {
	config.LoadConfig()
	mongoURI := config.GetEnv("MONGODB_URI", "")

	clientOptions := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(dbName)

	log.Printf("Successfully connected to the %s database!\n", dbName)

	return db, nil
}

func ConnectMongoClient() (*mongo.Client, error) {
	config.LoadConfig()
	mongoURI := config.GetEnv("MONGODB_URI", "")

	clientOptions := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("Failed to ping MongoDB: %v", err)
	}

	log.Printf("Successfully connected to the users and orders databases!\n")

	return client, nil
}
