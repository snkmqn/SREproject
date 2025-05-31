package repositories

import (
	"context"
	"user-service/internal/core/models"
	"user-service/internal/infrastructure/utils/security"
	"user-service/internal/interfaces/repositories"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type userRepositoryMongo struct {
	collection *mongo.Collection
	passwordHash security.PasswordHash
}

func NewUserRepositoryMongo(db *mongo.Database, passwordHash security.PasswordHash) repositories.UserRepository {
	return &userRepositoryMongo{
		collection:     db.Collection("users"),
		passwordHash: passwordHash,
	}
}

func (r *userRepositoryMongo) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		log.Println("Error while creating user:", err)
		return models.User{}, err
	}
	log.Println("User successfully created!")
	return user, nil
}

func (r *userRepositoryMongo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, nil
		}
		log.Println("Error while searching for user by email:", err)
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepositoryMongo) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, nil
		}
		log.Println("Error while searching for user by username:", err)
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepositoryMongo) AuthenticateUser(ctx context.Context, email, password string) (models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, errors.New("user not found")
		}
		log.Println("Error while searching for user by email:", err)
		return models.User{}, err
	}

	if !r.passwordHash.CheckPasswordHash(password, user.Password) {
		return models.User{}, errors.New("incorrect password")
	}

	return user, nil
}

func (r *userRepositoryMongo) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepositoryMongo) DeleteUser(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": userID})
	return err
}

