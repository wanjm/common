package common

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoQueryOperation struct {
	Collection  *mongo.Collection
	Context     context.Context
	filter      bson.M
	findOptions *options.FindOptions
}

func NewMongoQueryOperation(context context.Context, collection *mongo.Collection, opts []Optioner) *MongoQueryOperation {
	return &MongoQueryOperation{
		Collection:  collection,
		Context:     context,
		findOptions: options.Find(),
		filter:      GenMongoOption(opts),
	}
}

func (op *MongoQueryOperation) Count() (total int64, err error) {
	total, err = op.Collection.CountDocuments(op.Context, op.filter)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (op *MongoQueryOperation) Query(result any) (err error) {
	cur, err := op.Collection.Find(op.Context, op.filter, op.findOptions)
	if err != nil {
		return err
	}
	defer cur.Close(op.Context)
	err = cur.All(op.Context, result)
	if err != nil {
		return err
	}
	return err
}

func (op *MongoQueryOperation) SetSkip(skip int64) *options.FindOptions {
	op.findOptions.SetSkip(skip)
	return op.findOptions
}

func (op *MongoQueryOperation) SetLimit(limit int64) *options.FindOptions {
	op.findOptions.SetLimit(limit)
	return op.findOptions
}

func (op *MongoQueryOperation) SetProjection(cols []string) *options.FindOptions {
	projection := bson.M{}
	for _, col := range cols {
		projection[col] = 1
	}
	op.findOptions.SetProjection(projection)
	return op.findOptions
}
