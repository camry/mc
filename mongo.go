package mc

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClient struct {
    *mongo.Client
    ctx context.Context
}

func NewMongoClient(ctx context.Context, opts ...*options.ClientOptions) (*MongoClient, error) {
    c := &MongoClient{
        ctx: ctx,
    }
    client, err := mongo.Connect(c.ctx, opts...)
    if err != nil {
        return nil, err
    }
    err = client.Ping(c.ctx, readpref.Primary())
    if err != nil {
        return nil, err
    }
    c.Client = client
    return c, nil
}

func (x *MongoClient) Context() context.Context {
    return x.ctx
}

func (x *MongoClient) Database(name string, opts ...*options.DatabaseOptions) *MongoDatabase {
    return &MongoDatabase{
        ctx:      x.ctx,
        Database: x.Client.Database(name, opts...),
    }
}

func (x *MongoClient) Close() error {
    return x.Disconnect(x.ctx)
}
