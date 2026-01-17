package mc

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
)

type MongoCollection struct {
    *mongo.Collection
    ctx context.Context
}

func (x *MongoCollection) Context() context.Context {
    return x.ctx
}
