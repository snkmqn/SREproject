//go:build !integration
// +build !integration

package repositories

import (
	"context"
	"user-service/internal/core/models"
	"user-service/internal/interfaces/mongo"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCollection struct {
	mock.Mock
}

func (m *mockCollection) FindOne(ctx context.Context, filter interface{}) mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(mongo.SingleResult)
}

type mockSingleResult struct {
	mock.Mock
}

func (m *mockSingleResult) Decode(val interface{}) error {
	args := m.Called(val)
	return args.Error(0)
}

func TestGetUserByEmail(t *testing.T) {
	ctx := context.TODO()
	email := "test@example.com"
	expectedUser := models.User{ID: "123", Email: email}

	t.Run("User exists", func(t *testing.T) {
		collection := new(mockCollection)
		result := new(mockSingleResult)

		collection.On("FindOne", ctx, map[string]interface{}{"email": email}).Return(result)
		result.On("Decode", mock.Anything).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.User)
			*arg = expectedUser
		}).Return(nil)

		repo := &userRepositoryMongoMock{collection: collection}

		user, err := repo.GetUserByEmail(ctx, email)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("User not found", func(t *testing.T) {
		collection := new(mockCollection)
		result := new(mockSingleResult)

		collection.On("FindOne", ctx, map[string]interface{}{"email": email}).Return(result)
		result.On("Decode", mock.Anything).Return(errors.New("mongo: no documents in result"))

		repo := &userRepositoryMongoMock{collection: collection}

		user, err := repo.GetUserByEmail(ctx, email)

		assert.NoError(t, err)
		assert.Equal(t, models.User{}, user)
	})

	t.Run("Unexpected error", func(t *testing.T) {
		collection := new(mockCollection)
		result := new(mockSingleResult)

		dbErr := errors.New("db failure")
		collection.On("FindOne", ctx, map[string]interface{}{"email": email}).Return(result)
		result.On("Decode", mock.Anything).Return(dbErr)

		repo := &userRepositoryMongoMock{collection: collection}

		user, err := repo.GetUserByEmail(ctx, email)

		assert.Error(t, err)
		assert.Equal(t, models.User{}, user)
	})

}
