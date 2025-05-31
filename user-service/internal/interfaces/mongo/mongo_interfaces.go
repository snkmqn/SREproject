package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type Collection interface {
	FindOne(ctx context.Context, filter interface{}) SingleResult
}

type SingleResult interface {
	Decode(val interface{}) error
}

type MongoCollectionWrapper struct {
	Collection *mongo.Collection
}

func (m *MongoCollectionWrapper) FindOne(ctx context.Context, filter interface{}) SingleResult {
	return m.Collection.FindOne(ctx, filter)
}

type MongoSingleResultWrapper struct {
	Result *mongo.SingleResult
}

func (m *MongoSingleResultWrapper) Decode(val interface{}) error {
	return m.Result.Decode(val)
}
