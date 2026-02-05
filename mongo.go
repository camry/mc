package mc

import (
    "context"

    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoClient Mongo客户端结构体，封装了原生 *mongo.Client 客户端
type MongoClient struct {
    *mongo.Client                 // 继承原生 *mongo.Client 客户端
    ctx           context.Context // 上下文
}

// NewMongoClient 创建新的Mongo客户端
// ctx 上下文
// opts 客户端选项
func NewMongoClient(ctx context.Context, opts ...*options.ClientOptions) (*MongoClient, error) {
    c := &MongoClient{
        ctx: ctx, // 设置上下文
    }
    client, err := mongo.Connect(opts...) // 连接到 Mongo
    if err != nil {
        return nil, err
    }
    err = client.Ping(c.ctx, readpref.Primary()) // 测试连接
    if err != nil {
        return nil, err
    }
    c.Client = client // 设置原生客户端
    return c, nil
}

// Context 返回客户端的上下文
func (x *MongoClient) Context() context.Context {
    return x.ctx
}

// Database 获取指定名称的数据库实例
// name 数据库名称
// opts 数据库选项
func (x *MongoClient) Database(name string, opts ...options.Lister[options.DatabaseOptions]) *MongoDatabase {
    return &MongoDatabase{
        ctx:      x.ctx,                            // 使用相同的上下文
        Database: x.Client.Database(name, opts...), // 创建数据库实例
    }
}

// Close 关闭Mongo连接
func (x *MongoClient) Close() error {
    return x.Disconnect(x.ctx) // 断开连接
}
