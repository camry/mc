package mc

import (
    "context"
    "sync"

    "github.com/camry/g/v2/glog"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

type Collector struct {
    ctx                   context.Context             // ctx 上下文
    db                    *MongoDatabase              // db Mongo 数据库实例
    mu                    sync.RWMutex                // mu 读写锁
    closed                bool                        // closed 是否已关闭
    tableCollectorWorkers map[string]ICollectorWorker // tableCollectorWorkers 表名对应的收集器工作器
}

// NewCollector 创建新的收集器实例
// ctx 上下文
// db Mongo数据库实例
// opts 收集器选项
func NewCollector(ctx context.Context, db *MongoDatabase, opts ...CollectorOption) *Collector {
    x := &Collector{
        ctx:                   ctx,
        db:                    db,
        tableCollectorWorkers: make(map[string]ICollectorWorker),
    }
    for _, opt := range opts {
        opt(x)
    }
    return x
}

// Database 返回关联的Mongo数据库
func (x *Collector) Database() *MongoDatabase {
    return x.db
}

// Write 写入数据
// table 表名
// writeModel 写入模型
// opts 收集器工作选项，仅在该表的工作器首次创建时生效
func (x *Collector) Write(table string, writeModel mongo.WriteModel, opts ...CollectorWorkerOption) {
    worker, ok := x.getOrCreateWorker(table, opts...)
    if !ok {
        glog.Warnf("Collector Table %s Write After Closed", table)
        return
    }
    // 将 mongo.WriteModel 发送到工作器的通道（工作器若已被并发关闭，会丢弃并告警，不会 panic）
    worker.Write(writeModel)
}

// getOrCreateWorker 获取表对应的收集器工作器，不存在则创建并启动；收集器已关闭时返回 false
func (x *Collector) getOrCreateWorker(table string, opts ...CollectorWorkerOption) (ICollectorWorker, bool) {
    x.mu.RLock()
    worker, exists := x.tableCollectorWorkers[table]
    closed := x.closed
    x.mu.RUnlock()
    if exists {
        return worker, true
    }
    if closed {
        return nil, false
    }

    x.mu.Lock()
    defer x.mu.Unlock()
    if x.closed || x.ctx.Err() != nil {
        return nil, false
    }
    // 二次检查：并发首次写入同一张表时只创建一个工作器
    if worker, exists = x.tableCollectorWorkers[table]; exists {
        return worker, true
    }
    worker = NewCollectorWorker(x.ctx, x.db, table, opts...)
    x.tableCollectorWorkers[table] = worker
    go worker.Run() // 启动工作器协程
    return worker, true
}

// Close 关闭收集器，阻塞直到所有工作器的剩余数据写入完毕，可重复调用
// 须在关闭所有适配器之后调用，关闭后的写入会被丢弃
func (x *Collector) Close() error {
    x.mu.Lock()
    x.closed = true
    workers := make([]ICollectorWorker, 0, len(x.tableCollectorWorkers))
    for _, worker := range x.tableCollectorWorkers {
        workers = append(workers, worker)
    }
    x.tableCollectorWorkers = make(map[string]ICollectorWorker) // 清空映射
    x.mu.Unlock()

    // 并发关闭所有工作器
    var wg sync.WaitGroup
    for _, worker := range workers {
        wg.Go(worker.Close)
    }
    wg.Wait()
    return nil
}
