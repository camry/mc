package mc

import (
    "context"
    "strings"
    "sync"
    "time"

    "github.com/camry/g/v2/glog"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

// Adapter 适配器结构体，用于处理Mongo数据操作
// T 代表实现了IMessageTop接口的消息类型
// K 代表可比较的键类型
type Adapter[T IMessageTop, K BaseType] struct {
    t                     T                       // t 当前绑定的消息对象
    k                     K                       // k 当前绑定的键值
    ctx                   context.Context         // ctx 上下文
    cancel                context.CancelFunc      // cancel 上下文取消函数
    collector             ICollector              // collector 收集器实例
    collectorWorkerOption []CollectorWorkerOption // collectorWorkerOption 收集器工作选项
    buffer                chan IAdapterBsonM      // buffer 缓冲区通道
    bufferSize            int                     // bufferSize 缓冲区大小
    batchSize             int                     // batchSize 批量大小
    flushInterval         time.Duration           // flushInterval 刷新间隔
    mu                    sync.RWMutex            // mu 保护 closed 与 buffer 的关闭，写入持读锁，关闭持写锁
    closed                bool                    // closed 是否已关闭
    done                  chan struct{}           // done 后台协程退出信号（剩余数据已全部交给收集器）
}

// NewAdapter 创建新的适配器实例
// db Mongo数据库实例
// opts 适配器选项
func NewAdapter[T IMessageTop, K BaseType](opts ...AdapterOption[T, K]) *Adapter[T, K] {
    x := &Adapter[T, K]{
        ctx:           context.Background(),
        bufferSize:    1000,
        batchSize:     10,
        flushInterval: 10 * time.Millisecond,
    }
    for _, opt := range opts {
        opt(x)
    }
    ctxWithCancel, cancel := context.WithCancel(x.ctx)
    x.ctx = ctxWithCancel
    x.cancel = cancel
    if x.collector != nil {
        x.buffer = make(chan IAdapterBsonM, x.bufferSize)
        x.done = make(chan struct{})
        go x.writeCollectorWorker()
    }
    context.AfterFunc(x.ctx, x.shutdown) // 外部上下文取消时同样停止接收写入
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
// nestedMessages 嵌套消息列表
// key 字段键
// val 字段值
func (x *Adapter[T, K]) Set(nestedMessages []IMessage, key, val any) {
    if x.collector == nil {
        glog.Errorf("Adapter Table %s Set Failed Collector Not Initialized", x.t.Table())
        return
    }
    mk := toNestedMessageFieldKey(nestedMessages, key)           // 构建嵌套消息字段键
    x.write(newAdapterBsonM(AggregateStageSet, bson.M{mk: val})) // 使用$set操作符设置字段值
    glog.Debugf("Adapter Table %s Set _id=%v FieldKey=%s FieldValue=%v", x.t.Table(), x.GetKey(), mk, val)
}

// Unset 取消设置字段值
// nestedMessages 嵌套消息列表
// key 要取消设置的字段键
func (x *Adapter[T, K]) Unset(nestedMessages []IMessage, key any) {
    if x.collector == nil {
        glog.Errorf("Adapter Table %s Unset Failed Collector Not Initialized", x.t.Table())
        return
    }
    mk := toNestedMessageFieldKey(nestedMessages, key)            // 构建嵌套消息字段键
    x.write(newAdapterBsonM(AggregateStageUnset, bson.M{mk: ""})) // 使用$unset操作符取消设置字段值
    glog.Debugf("Adapter Table %s Unset FieldKey=%s", x.t.Table(), mk)
}

// write 写入数据到缓冲区
// 持读锁发送：关闭方需持写锁才能 close(buffer)，因此不会向已关闭的通道发送；
// 缓冲区满时阻塞等待后台协程消费，后台协程不依赖该锁，不会死锁。
func (x *Adapter[T, K]) write(abm IAdapterBsonM) {
    x.mu.RLock()
    defer x.mu.RUnlock()
    if x.closed || x.ctx.Err() != nil { // AfterFunc 异步执行 shutdown，取消后立即拒绝写入
        glog.Warnf("Adapter Table %s Write After Closed", x.t.Table())
        return
    }
    x.buffer <- abm
}

// shutdown 停止接收写入并关闭缓冲区通道，可重复调用
func (x *Adapter[T, K]) shutdown() {
    x.mu.Lock()
    defer x.mu.Unlock()
    if x.closed {
        return
    }
    x.closed = true
    if x.buffer != nil {
        close(x.buffer) // 后台协程消费完剩余数据后退出
    }
}

// writeCollectorWorker 写入数据到收集器工作器，缓冲区通道关闭且剩余数据处理完毕后退出
func (x *Adapter[T, K]) writeCollectorWorker() {
    defer close(x.done)
    ticker := time.NewTicker(x.flushInterval)
    defer ticker.Stop()
    batchAdapterBsonM := make([]IAdapterBsonM, 0, x.batchSize)
    for {
        select {
        case data, ok := <-x.buffer:
            if !ok {
                // 通道已关闭，处理剩余数据
                if len(batchAdapterBsonM) > 0 {
                    x.doBatchAdapterBsonM(batchAdapterBsonM)
                }
                glog.Debugf("Adapter writeCollectorWorker Table %s Closed", x.t.Table())
                return
            }
            batchAdapterBsonM = append(batchAdapterBsonM, data)
            if len(batchAdapterBsonM) >= x.batchSize {
                x.doBatchAdapterBsonM(batchAdapterBsonM)
                batchAdapterBsonM = batchAdapterBsonM[:0]
            }
        case <-ticker.C:
            if len(batchAdapterBsonM) > 0 {
                x.doBatchAdapterBsonM(batchAdapterBsonM)
                batchAdapterBsonM = batchAdapterBsonM[:0]
                glog.Debugf("Adapter writeCollectorWorker Table %s Flush Interval", x.t.Table())
            }
        }
    }
}

// doBatchAdapterBsonM 批量处理 adapterBsonM 列表
func (x *Adapter[T, K]) doBatchAdapterBsonM(batchAdapterBsonMList []IAdapterBsonM) {
    if len(batchAdapterBsonMList) == 0 {
        return
    }
    var currentAggregateStage AggregateStage
    currentM := make(bson.M)
    flush := func() {
        if len(currentM) == 0 {
            return
        }
        switch currentAggregateStage {
        case AggregateStageSet:
            x.collector.Write(x.t.Table(), mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": x.GetKey()}).SetUpdate(bson.M{
                "$set": currentM,
            }), x.collectorWorkerOption...)
            glog.Debugf("Adapter doBatchAdapterBsonM Table %s Set _id=%v Field=%v", x.t.Table(), x.GetKey(), currentM)
        case AggregateStageUnset:
            x.collector.Write(x.t.Table(), mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": x.GetKey()}).SetUpdate(bson.M{
                "$unset": currentM,
            }), x.collectorWorkerOption...)
            glog.Debugf("Adapter doBatchAdapterBsonM Table %s Unset _id=%v Field=%v", x.t.Table(), x.GetKey(), currentM)
        }
        // 重置
        currentM = make(bson.M)
    }
    for _, abm := range batchAdapterBsonMList {
        // 如果阶段变化，先 flush 上一组
        if abm.AggregateStage() != currentAggregateStage && len(currentM) > 0 {
            flush()
        }
        currentAggregateStage = abm.AggregateStage()
        // 合并 bson.M
        for k, v := range abm.M() {
            hasConflict := false
            for currentK := range currentM {
                // 相同键直接覆盖，跳过检测
                if k == currentK {
                    break
                }
                // 检测父子键关系：currentK 是 k 的子路径 或 k 是 currentK 的子路径
                if strings.HasPrefix(currentK, k+".") || strings.HasPrefix(k, currentK+".") {
                    hasConflict = true
                    break
                }
            }
            if hasConflict {
                flush()
            }
            currentM[k] = v
        }
    }
    // 处理最后一组
    flush()
}

// Close 关闭适配器，阻塞直到剩余数据全部交给收集器，可重复调用
// 须在关闭收集器之前调用，否则剩余数据会被已关闭的收集器丢弃
func (x *Adapter[T, K]) Close() error {
    x.cancel()
    x.shutdown()
    if x.done != nil {
        <-x.done
    }
    return nil
}
