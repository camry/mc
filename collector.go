package mc

import (
    "context"
    "sync"

    "go.mongodb.org/mongo-driver/mongo"
)

type Collector struct {
    ctx                   context.Context             // ctx 上下文
    db                    *MongoDatabase              // db Mongo 数据库实例
    mu                    sync.RWMutex                // mu 读写锁
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

// Write 写入数据
// table 表名
// writeModel 写入模型
// opts 收集器工作选项
func (x *Collector) Write(table string, writeModel mongo.WriteModel, opts ...CollectorWorkerOption) {
    // 检查是否存在对应表的收集器工作器
    x.mu.RLock()
    worker, exists := x.tableCollectorWorkers[table]
    x.mu.RUnlock()
    if !exists {
        x.mu.Lock()
        // 如果不存在，则创建一个新的收集器工作器
        worker = NewCollectorWorker(x.ctx, x.db, table, opts...)
        x.tableCollectorWorkers[table] = worker
        // 启动工作器协程
        go worker.Run()
        x.mu.Unlock()
    }

    // 将 mongo.WriteModel 发送到工作器的通道
    worker.Write(writeModel)
}

// Close 关闭收集器
func (x *Collector) Close() error {
    x.mu.Lock()
    workers := make([]ICollectorWorker, 0, len(x.tableCollectorWorkers))
    for _, worker := range x.tableCollectorWorkers {
        workers = append(workers, worker)
    }
    x.tableCollectorWorkers = make(map[string]ICollectorWorker) // 清空映射
    x.mu.Unlock()

    // 并发关闭所有工作器
    var wg sync.WaitGroup
    wg.Add(len(workers))
    for _, worker := range workers {
        go func(w ICollectorWorker) {
            defer wg.Done()
            w.Close()
        }(worker)
    }
    wg.Wait()
    return nil
}
