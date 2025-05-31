package repositories

import (
	"context"
	"user-service/internal/core/models"
	"user-service/internal/interfaces/mongo"
)

type userRepositoryMongoMock struct {
	collection mongo.Collection
}

func (r *userRepositoryMongoMock) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, map[string]interface{}{"email": email}).Decode(&user)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return models.User{}, nil
		}
		return models.User{}, err
	}
	return user, nil
}
