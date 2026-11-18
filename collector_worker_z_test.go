package mc

import (
    "context"
    "errors"
    "testing"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

func newTestWorker() *CollectorWorker {
    return &CollectorWorker{
        ctx:           context.Background(),
        table:         "test",
        maxRetries:    2,
        retryBaseDelay: 0,
        retryMaxDelay:  0,
    }
}

func newTestModels(n int) []mongo.WriteModel {
    models := make([]mongo.WriteModel, 0, n)
    for i := range n {
        models = append(models, mongo.NewUpdateOneModel().
            SetFilter(bson.M{"_id": i}).
            SetUpdate(bson.M{"$set": bson.M{"v": i}}))
    }
    return models
}

func TestSplitBulkWriteErrors(t *testing.T) {
    models := newTestModels(3)

    bulkErr := mongo.BulkWriteException{WriteErrors: []mongo.BulkWriteError{
        {WriteError: mongo.WriteError{Index: 1, Code: 40, Message: "conflicting"}},
    }}
    remaining, dropped, ok := splitBulkWriteErrors(models, bulkErr)
    if !ok {
        t.Fatal("expected ok=true for BulkWriteException with valid index")
    }
    if len(dropped) != 1 || dropped[0] != models[1] {
        t.Fatalf("expected dropped=[models[1]], got %v", dropped)
    }
    if len(remaining) != 2 || remaining[0] != models[0] || remaining[1] != models[2] {
        t.Fatalf("unexpected remaining %v", remaining)
    }

    // 非 BulkWriteException 不拆分
    if _, _, ok = splitBulkWriteErrors(models, errors.New("network")); ok {
        t.Fatal("expected ok=false for plain error")
    }
    // 下标越界不拆分
    badIndex := mongo.BulkWriteException{WriteErrors: []mongo.BulkWriteError{
        {WriteError: mongo.WriteError{Index: 9}},
    }}
    if _, _, ok = splitBulkWriteErrors(models, badIndex); ok {
        t.Fatal("expected ok=false for out-of-range index")
    }
}

func TestDoBatchWriteModelIsolatesPoison(t *testing.T) {
    w := newTestWorker()
    models := newTestModels(3)

    var calls [][]mongo.WriteModel
    w.bulkWriteFn = func(batch []mongo.WriteModel) error {
        calls = append(calls, append([]mongo.WriteModel(nil), batch...))
        if len(calls) == 1 {
            // 第一条是毒丸：服务端拒绝对应下标 1
            return mongo.BulkWriteException{WriteErrors: []mongo.BulkWriteError{
                {WriteError: mongo.WriteError{Index: 1, Code: 40, Message: "conflicting"}},
            }}
        }
        return nil
    }

    w.doBatchWriteModel(models)

    if len(calls) != 2 {
        t.Fatalf("expected 2 bulkWrite calls, got %d", len(calls))
    }
    if len(calls[1]) != 2 || calls[1][0] != models[0] || calls[1][1] != models[2] {
        t.Fatalf("expected remaining models [0,2], got %d models", len(calls[1]))
    }
}

func TestDoBatchWriteModelRetryExhausted(t *testing.T) {
    w := newTestWorker()
    models := newTestModels(2)

    calls := 0
    w.bulkWriteFn = func([]mongo.WriteModel) error {
        calls++
        return errors.New("network down")
    }

    w.doBatchWriteModel(models)

    // 首次 + maxRetries 次重试
    if calls != w.maxRetries+1 {
        t.Fatalf("expected %d calls, got %d", w.maxRetries+1, calls)
    }
}
