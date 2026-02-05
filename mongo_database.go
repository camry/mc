package mc

import (
    "bytes"
    "context"
    "errors"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoDatabase Mongo数据库结构体，封装了原生 *mongo.Database 数据库
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
func (x *MongoDatabase) Collection(name string, opts ...options.Lister[options.CollectionOptions]) *MongoCollection {
    return &MongoCollection{
        ctx:        x.ctx,                                // 使用相同的上下文
        Collection: x.Database.Collection(name, opts...), // 创建集合实例
    }
}

// Find 查找指定顶级消息对象
func (x *MongoDatabase) Find(top IMessageTop) error {
    decode, err := GetDecode(top)
    if err != nil {
        return err
    }
    bsonRaw, err := x.Collection(top.Table()).FindOne(x.Context(), bson.M{"_id": top.Key()}).Raw()
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            _, err = x.Collection(top.Table()).InsertOne(x.Context(), top.Marshal())
            if err != nil {
                return err
            }
            return nil
        }
        return err
    }
    err = decode.Decode(bson.NewDocumentReader(bytes.NewReader(bsonRaw)), top)
    if err != nil {
        return err
    }
    return nil
}
