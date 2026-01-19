package mc

import (
    "time"

    "github.com/camry/g/glog"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type Adapter[T IMessageTop, K comparable] struct {
    db      *MongoDatabase
    t       T
    k       K
    setData chan bson.M
}

func NewAdapter[T IMessageTop, K comparable](db *MongoDatabase) *Adapter[T, K] {
    x := &Adapter[T, K]{
        db:      db,
        setData: make(chan bson.M, 1000),
    }
    go x.Consume()
    return x
}

func (x *Adapter[T, K]) Bind(t T, k K) {
    x.t = t
    x.k = k
}

func (x *Adapter[T, K]) GetMessageTop() T {
    return x.t
}

func (x *Adapter[T, K]) GetKey() K {
    return x.k
}

func (x *Adapter[T, K]) Set(topMessage IMessageTop, nestedMessages []IMessage, key, val any) {
    mk := toNestedMessageFieldKey(nestedMessages, key)
    x.setData <- bson.M{
        "$set": bson.M{
            mk: val,
        },
    }
    glog.Debugf("Adapter Set Table=%s _id=%v FieldKey:%s", x.t.Table(), x.GetKey(), mk)
}

func (x *Adapter[T, K]) Unset(topMessage IMessageTop, nestedMessages []IMessage, key any) {
    glog.Debugf("Adapter Unset Field Key:%s", toNestedMessageFieldKey(nestedMessages, key))
}

func (x *Adapter[T, K]) Consume() {
    const batchSize = 100
    const flushInterval = 100 * time.Millisecond

    batchM := make([]bson.M, 0, batchSize)
    ticker := time.NewTicker(flushInterval)
    defer ticker.Stop()

    for {
        select {
        case data := <-x.setData:
            batchM = append(batchM, data)
            if len(batchM) >= batchSize {
                x.flushBatch(batchM)
                batchM = batchM[:0]
            }
        case <-ticker.C:
            if len(batchM) > 0 {
                x.flushBatch(batchM)
                batchM = batchM[:0]
            }
        }
    }
}

// flushBatch 执行批量写入操作
func (x *Adapter[T, K]) flushBatch(batchM []bson.M) {
    if len(batchM) == 0 {
        return
    }
    var writeModels []mongo.WriteModel
    for _, update := range batchM {
        writeModel := mongo.NewUpdateOneModel().
            SetFilter(bson.M{"_id": x.GetKey()}).
            SetUpdate(update)
        writeModels = append(writeModels, writeModel)
    }
    _, err := x.db.Collection(x.t.Table()).BulkWrite(x.db.Context(), writeModels)
    if err != nil {
        glog.Errorf("Adapter Consume BulkWrite Error: %v", err)
    }
}
