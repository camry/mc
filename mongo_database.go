package mc

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDatabase struct {
    *mongo.Database
    ctx context.Context
}

func (x *MongoDatabase) Context() context.Context {
    return x.ctx
}

func (x *MongoDatabase) Collection(name string, opts ...*options.CollectionOptions) *MongoCollection {
    return &MongoCollection{
        ctx:        x.ctx,
        Collection: x.Database.Collection(name, opts...),
    }
}
