package mc

import (
    "time"

    "github.com/camry/g/glog"
    "github.com/samber/lo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

// Adapter 适配器结构体，用于处理MongoDB数据操作
// T 代表实现了IMessageTop接口的消息类型
// K 代表可比较的键类型
type Adapter[T IMessageTop, K comparable] struct {
    t          T              // t 当前绑定的消息对象
    k          K              // k 当前绑定的键值
    db         *MongoDatabase // db MongoDB数据库实例
    enableSave bool           // enableSave 保存数据
    size       int            // size 缓冲容量
    setData    chan bson.M    // setData 用于接收待设置数据的通道
}

// NewAdapter 创建新的适配器实例
// db MongoDB数据库实例
func NewAdapter[T IMessageTop, K comparable](db *MongoDatabase, opts ...AdapterOption[T, K]) *Adapter[T, K] {
    x := &Adapter[T, K]{
        db: db,
    }
    for _, opt := range opts {
        opt(x)
    }
    if x.enableSave {
        x.setData = make(chan bson.M, lo.Max([]int{x.size, 100}))
        go x.Consume()
    }
    return x
}

// Bind 绑定消息对象和键值
// t 顶级消息对象
// k 键值
func (x *Adapter[T, K]) Bind(t T, k K) {
    x.t = t
    x.k = k
}

// GetMessageTop 获取绑定的顶级消息对象
func (x *Adapter[T, K]) GetMessageTop() T {
    return x.t
}

// GetKey 获取绑定的键值
func (x *Adapter[T, K]) GetKey() K {
    return x.k
}

// Set 设置字段值
// topMessage 顶级消息对象
// nestedMessages 嵌套消息列表
// key 字段键
// val 字段值
func (x *Adapter[T, K]) Set(topMessage IMessageTop, nestedMessages []IMessage, key, val any) {
    if x.enableSave {
        mk := toNestedMessageFieldKey(nestedMessages, key) // 构建嵌套消息字段键
        x.setData <- bson.M{
            "$set": bson.M{
                mk: val, // 使用$set操作符设置字段值
            },
        }
        glog.Debugf("Adapter Set Table=%s _id=%v FieldKey:%s FieldValue:%v", x.t.Table(), x.GetKey(), mk, val)
    }
}

// Unset 取消设置字段值
// topMessage 顶级消息对象
// nestedMessages 嵌套消息列表
// key 要取消设置的字段键
func (x *Adapter[T, K]) Unset(topMessage IMessageTop, nestedMessages []IMessage, key any) {
    if x.enableSave {
        mk := toNestedMessageFieldKey(nestedMessages, key) // 构建嵌套消息字段键
        x.setData <- bson.M{
            "$unset": bson.M{
                mk: "", // 使用$unset操作符取消设置字段值
            },
        }
        glog.Debugf("Adapter Unset Field Key:%s", mk)
    }
}

// Consume 消费数据通道中的更新请求，批量执行数据库操作
func (x *Adapter[T, K]) Consume() {
    if !x.enableSave {
        return
    }

    const batchSize = 100                        // 批量处理大小
    const flushInterval = 100 * time.Millisecond // 刷新间隔

    batchM := make([]bson.M, 0, batchSize)  // 批量数据切片
    ticker := time.NewTicker(flushInterval) // 定时器
    defer ticker.Stop()

    for {
        select {
        case data := <-x.setData: // 接收数据设置请求
            batchM = append(batchM, data)
            if len(batchM) >= batchSize { // 达到批量大小则执行批量操作
                x.flushBatch(batchM)
                batchM = batchM[:0] // 重置切片
            }
        case <-ticker.C: // 定时刷新
            if len(batchM) > 0 {
                x.flushBatch(batchM)
                batchM = batchM[:0] // 重置切片
            }
        }
    }
}

// flushBatch 执行批量写入操作
// batchM 批量数据
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
