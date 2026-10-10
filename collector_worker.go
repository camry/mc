package mc

import (
    "context"
    "sync"
    "time"

    "github.com/camry/g/v2/glog"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

type CollectorWorker struct {
    ctx           context.Context       // ctx 上下文
    cancel        context.CancelFunc    // cancel 上下文取消函数
    db            *MongoDatabase        // db Mongo 数据库实例
    table         string                // table 表名
    buffer        chan mongo.WriteModel // buffer 缓冲区通道
    bufferSize    int                   // bufferSize 缓冲区大小
    batchSize     int                   // batchSize 批量大小
    flushInterval time.Duration         // flushInterval 刷新间隔
    writeTimeout  time.Duration         // writeTimeout 单次 BulkWrite 超时时间，<=0 表示不限制
    mu            sync.RWMutex          // mu 保护 closed 与 buffer 的关闭，写入持读锁，关闭持写锁
    closed        bool                  // closed 是否已关闭
    done          chan struct{}         // done Run 协程退出信号（剩余数据已全部写入）
}

// NewCollectorWorker 创建新的收集器工作实例，创建后须调用 Run 启动
func NewCollectorWorker(ctx context.Context, db *MongoDatabase, table string, opts ...CollectorWorkerOption) *CollectorWorker {
    ctxWithCancel, cancel := context.WithCancel(ctx)
    x := &CollectorWorker{
        ctx:           ctxWithCancel,
        cancel:        cancel,
        db:            db,
        table:         table,
        bufferSize:    10000,
        batchSize:     100,
        flushInterval: 100 * time.Millisecond,
        writeTimeout:  10 * time.Second,
        done:          make(chan struct{}),
    }
    for _, opt := range opts {
        opt(x)
    }
    x.buffer = make(chan mongo.WriteModel, x.bufferSize)
    context.AfterFunc(x.ctx, x.shutdown) // 外部上下文取消时同样停止接收写入
    return x
}

// Write 写入数据到缓冲区
// 持读锁发送：关闭方需持写锁才能 close(buffer)，因此不会向已关闭的通道发送；
// 缓冲区满时阻塞等待 Run 协程消费，Run 协程不依赖该锁，不会死锁。
func (x *CollectorWorker) Write(writeModel mongo.WriteModel) {
    x.mu.RLock()
    defer x.mu.RUnlock()
    if x.closed || x.ctx.Err() != nil { // AfterFunc 异步执行 shutdown，取消后立即拒绝写入
        glog.Warnf("CollectorWorker Table %s Write After Closed", x.table)
        return
    }
    x.buffer <- writeModel
}

// shutdown 停止接收写入并关闭缓冲区通道，可重复调用
func (x *CollectorWorker) shutdown() {
    x.mu.Lock()
    defer x.mu.Unlock()
    if x.closed {
        return
    }
    x.closed = true
    close(x.buffer) // Run 协程消费完剩余数据后退出
}

// Run 启动工作器，缓冲区通道关闭且剩余数据写入完毕后退出
func (x *CollectorWorker) Run() {
    defer close(x.done)
    ticker := time.NewTicker(x.flushInterval)
    defer ticker.Stop()
    batchWriteModels := make([]mongo.WriteModel, 0, x.batchSize)
    for {
        select {
        case data, ok := <-x.buffer:
            if !ok {
                // 通道已关闭，处理剩余数据
                if len(batchWriteModels) > 0 {
                    x.doBatchWriteModel(batchWriteModels)
                }
                glog.Debugf("CollectorWorker Table %s Run Closed", x.table)
                return
            }
            batchWriteModels = append(batchWriteModels, data)
            if len(batchWriteModels) >= x.batchSize {
                x.doBatchWriteModel(batchWriteModels)
                batchWriteModels = batchWriteModels[:0]
            }
        case <-ticker.C:
            if len(batchWriteModels) > 0 {
                x.doBatchWriteModel(batchWriteModels)
                batchWriteModels = batchWriteModels[:0]
                glog.Debugf("CollectorWorker Table %s Run Flush Interval", x.table)
            }
        }
    }
}

// Close 关闭工作器，阻塞直到剩余数据全部写入，可重复调用
func (x *CollectorWorker) Close() {
    x.cancel()
    x.shutdown()
    <-x.done
}

// doBatchWriteModel 执行批量写入
func (x *CollectorWorker) doBatchWriteModel(batchWriteModels []mongo.WriteModel) {
    if len(batchWriteModels) == 0 {
        return
    }
    ctx := context.WithoutCancel(x.db.Context())
    if x.writeTimeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, x.writeTimeout)
        defer cancel()
    }
    _, err := x.db.Collection(x.table).BulkWrite(ctx, batchWriteModels)
    if err != nil {
        glog.Errorf("CollectorWorker Table %s BulkWrite Error: %v", x.table, err)
    }
}
