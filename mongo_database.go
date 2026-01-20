package mc

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDatabase MongoDB数据库结构体，封装了原生 *mongo.Database 数据库
type MongoDatabase struct {
    *mongo.Database                 // 继承原生 *mongo.Database 数据库
    ctx             context.Context // 上下文
}

// Context 返回数据库的上下文
func (x *MongoDatabase) Context() context.Context {
    return x.ctx
}

// Collection 获取指定名称的集合实例
// name 集合名称
// opts 集合选项
func (x *MongoDatabase) Collection(name string, opts ...*options.CollectionOptions) *MongoCollection {
    return &MongoCollection{
        ctx:        x.ctx,                                // 使用相同的上下文
        Collection: x.Database.Collection(name, opts...), // 创建集合实例
    }
}
