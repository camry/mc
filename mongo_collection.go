package mc

import (
    "context"

    "go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoCollection Mongo集合结构体，封装了原生 *mongo.Collection 集合
type MongoCollection struct {
    *mongo.Collection                 // 继承原生 *mongo.Collection 集合
    ctx               context.Context // 上下文
}

// Context 返回集合的上下文
func (x *MongoCollection) Context() context.Context {
    return x.ctx
}
