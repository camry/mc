package mc

import (
    "bytes"
    "reflect"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoTable[T IMessageTop, K BaseType] struct {
    db *MongoDatabase
}

func NewMongoTable[T IMessageTop, K BaseType](db *MongoDatabase) *MongoTable[T, K] {
    return &MongoTable[T, K]{
        db: db,
    }
}

func (x *MongoTable[T, K]) FindOne(id K) (T, error) {
    var zero T

    t := reflect.TypeFor[T]().Elem()
    top := reflect.New(t).Interface().(T)
    decode, err := GetDecode(top)
    if err != nil {
        return zero, err
    }

    bsonRaw, err := x.db.Collection(top.Table()).FindOne(x.db.Context(), bson.M{"_id": id}).Raw()
    if err != nil {
        return zero, err
    }
    err = decode.Decode(bson.NewDocumentReader(bytes.NewReader(bsonRaw)), top)
    if err != nil {
        return zero, err
    }

    return top, nil
}

func (x *MongoTable[T, K]) FindMany(ids []K) ([]T, error) {
    t := reflect.TypeFor[T]().Elem()
    top := reflect.New(t).Interface().(T)
    decode, err := GetDecode(top)
    if err != nil {
        return nil, err
    }

    cursor, err := x.db.Collection(top.Table()).Find(x.db.Context(), bson.M{"_id": bson.M{"$in": ids}}, options.Find().SetBatchSize(2048))
    if err != nil {
        return nil, err
    }
    defer cursor.Close(x.db.Context())

    lst := make([]T, 0, cursor.RemainingBatchLength())
    for cursor.Next(x.db.Context()) {
        tt := reflect.New(t).Interface().(T)
        err = decode.Decode(bson.NewDocumentReader(bytes.NewReader(cursor.Current)), tt)
        if err != nil {
            return nil, err
        }
        lst = append(lst, tt)
    }

    return lst, nil
}
