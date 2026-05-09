package mc

import (
    "context"
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
}

// NewCollectorWorker 创建新的收集器工作实例
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
    }
    for _, opt := range opts {
        opt(x)
    }
    x.buffer = make(chan mongo.WriteModel, x.bufferSize)
    return x
}

// Write 写入数据到缓冲区
func (x *CollectorWorker) Write(writeModel mongo.WriteModel) {
    select {
    case x.buffer <- writeModel:
    case <-x.ctx.Done():
        // 上下文已完成，无法发送
        glog.Warnf("CollectorWorker Table %s Write Context Done", x.table)
    }
}

// Run 启动工作器
func (x *CollectorWorker) Run() {
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
        case <-x.ctx.Done():
            // 关闭缓冲区通道，让剩余数据被处理
            close(x.buffer)
            // 继续处理缓冲区中的剩余数据直到通道关闭
            for data := range x.buffer {
                batchWriteModels = append(batchWriteModels, data)
                if len(batchWriteModels) >= x.batchSize {
                    x.doBatchWriteModel(batchWriteModels)
                    batchWriteModels = batchWriteModels[:0]
                }
            }
            // 处理最后一批数据
            if len(batchWriteModels) > 0 {
                x.doBatchWriteModel(batchWriteModels)
            }
            glog.Debugf("CollectorWorker Table %s Run Context Done", x.table)
            return
        }
    }
}

// Close 关闭工作器
func (x *CollectorWorker) Close() {
    x.cancel()
}

// doBatchWriteModel 执行批量写入
func (x *CollectorWorker) doBatchWriteModel(batchWriteModels []mongo.WriteModel) {
    if len(batchWriteModels) == 0 {
        return
    }
    _, err := x.db.Collection(x.table).BulkWrite(x.db.Context(), batchWriteModels)
    if err != nil {
        glog.Errorf("CollectorWorker Table %s BulkWrite Error: %v", x.table, err)
    }
}
