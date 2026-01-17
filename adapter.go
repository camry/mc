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

func (a *Adapter[T, K]) Bind(t T, k K) {
    a.t = t
    a.k = k
}

func (a *Adapter[T, K]) GetMessageTop() T {
    return a.t
}

func (a *Adapter[T, K]) GetKey() K {
    return a.k
}

func (a *Adapter[T, K]) Set(topMessage IMessageTop, nestedMessages []IMessage, key, val any) {
    mk := toNestedMessageFieldKey(nestedMessages, key)
    a.setData <- bson.M{
        "$set": bson.M{
            mk: val,
        },
    }
    glog.Debugf("Adapter Set Table=%s _id=%v FieldKey:%s", a.t.Table(), a.GetKey(), mk)
}

func (a *Adapter[T, K]) Unset(topMessage IMessageTop, nestedMessages []IMessage, key any) {
    glog.Debugf("Adapter Unset Field Key:%s", toNestedMessageFieldKey(nestedMessages, key))
}

func (a *Adapter[T, K]) Consume() {
    const batchSize = 100
    const flushInterval = 100 * time.Millisecond

    batchM := make([]bson.M, 0, batchSize)
    ticker := time.NewTicker(flushInterval)
    defer ticker.Stop()

    for {
        select {
        case data := <-a.setData:
            batchM = append(batchM, data)
            if len(batchM) >= batchSize {
                a.flushBatch(batchM)
                batchM = batchM[:0]
            }
        case <-ticker.C:
            if len(batchM) > 0 {
                a.flushBatch(batchM)
                batchM = batchM[:0]
            }
        }
    }
}

// flushBatch 执行批量写入操作
func (a *Adapter[T, K]) flushBatch(batchM []bson.M) {
    if len(batchM) == 0 {
        return
    }
    var writeModels []mongo.WriteModel
    for _, update := range batchM {
        writeModel := mongo.NewUpdateOneModel().
            SetFilter(bson.M{"_id": a.GetKey()}).
            SetUpdate(update)
        writeModels = append(writeModels, writeModel)
    }
    _, err := a.db.Collection(a.t.Table()).BulkWrite(a.db.Context(), writeModels)
    if err != nil {
        glog.Errorf("Adapter Consume BulkWrite Error: %v", err)
    }
}
