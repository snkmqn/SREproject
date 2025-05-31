package repositories

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"user-service/internal/core/models"
	"user-service/internal/interfaces/repositories"
)

type ProductRepositoryMongo struct {
	collection         *mongo.Collection
	categoryCollection *mongo.Collection
	Client             *mongo.Client
}

func NewProductRepositoryMongo(db *mongo.Database) repositories.ProductRepository {
	if db == nil {
		log.Println("mongo database is not initialized")
		return nil
	}
	productCol := db.Collection("products")
	categoryCol := db.Collection("categories")

	if productCol == nil || categoryCol == nil {
		log.Println("failed to initialize MongoDB collections")
		return nil
	}

	return &ProductRepositoryMongo{
		collection:         productCol,
		categoryCollection: categoryCol,
	}
}

func (r *ProductRepositoryMongo) CreateProduct(ctx context.Context, product models.Product) (models.Product, error) {

	var existingProduct models.Product
	log.Printf("Checking if product exists with name: %s and category_id: %s", product.Name, product.CategoryID)

	err := r.collection.FindOne(ctx, bson.M{"name": product.Name, "category_id": product.CategoryID}).Decode(&existingProduct)

	if err == nil {
		log.Printf("Product with name %s already exists", product.Name)
		return models.Product{}, fmt.Errorf("Product already exists")
	}

	_, err = r.collection.InsertOne(ctx, product)
	if err != nil {
		log.Printf("Error inserting product: %v", err)
		return models.Product{}, err
	}
	log.Printf("Product %s created successfully", product.Name)

	return product, nil
}

func (r *ProductRepositoryMongo) GetProductByID(ctx context.Context, id string) (models.Product, error) {
	var product models.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (r *ProductRepositoryMongo) UpdateProduct(ctx context.Context, id string, product models.Product) (models.Product, error) {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": product})
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (r *ProductRepositoryMongo) DeleteProduct(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *ProductRepositoryMongo) getCategoryIDByName(ctx context.Context, name string) (string, error) {
	var category struct {
		ID   string `bson:"_id"`
		Name string `bson:"name"`
	}
	err := r.categoryCollection.FindOne(ctx, bson.M{"name": name}).Decode(&category)
	if err != nil {
		return "", err
	}
	return category.ID, nil
}

func (r *ProductRepositoryMongo) ListProducts(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]models.Product, error) {
	if categoryName, ok := filter["category_name"]; ok {
		categoryID, err := r.getCategoryIDByName(ctx, fmt.Sprintf("%v", categoryName))
		if err != nil {
			log.Printf("Category '%v' not found: %v", categoryName, err)
			return nil, mongo.ErrNoDocuments
		}
		filter["category_id"] = categoryID
		delete(filter, "category_name")
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSkip(skip).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		if product.ID != "" {
			if _, err := uuid.Parse(product.ID); err != nil {
				log.Printf("Error parsing UUID from _id: %v", err)
				return nil, fmt.Errorf("failed to parse UUID from _id: %v", err)
			}
		}
		products = append(products, product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return products, nil
}

func (r *ProductRepositoryMongo) DecreaseStock(ctx context.Context, productID string, quantity int) (models.Product, error) {
	log.Printf("Attempting to decrease stock for product %s by %d", productID, quantity)

	var product models.Product

	err := r.collection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		return models.Product{}, fmt.Errorf("product not found: %w", err)
	}

	if product.Stock < quantity {
		return models.Product{}, fmt.Errorf("insufficient stock")
	}

	update := bson.M{
		"$inc": bson.M{"stock": -quantity},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": productID}, update)
	if err != nil {
		return models.Product{}, fmt.Errorf("failed to decrease stock: %w", err)
	}

	if result.MatchedCount == 0 {
		log.Printf("No matching product found for ID %s", productID)
	}

	if result.ModifiedCount == 0 {
		log.Printf("Stock not decreased for product %s", productID)
	} else {
		log.Printf("Stock decreased successfully for product %s", productID)
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		return models.Product{}, fmt.Errorf("failed to retrieve updated product: %w", err)
	}

	return product, nil
}
