package mc

import "time"

// CollectorWorkerOption 收集器工作选项
type CollectorWorkerOption func(x *CollectorWorker)

// WithCollectorWorkerBufferSize 缓冲区大小
func WithCollectorWorkerBufferSize(bufferSize int) CollectorWorkerOption {
    return func(x *CollectorWorker) {
        x.bufferSize = bufferSize
    }
}

// WithCollectorWorkerBatchSize 批量大小
func WithCollectorWorkerBatchSize(batchSize int) CollectorWorkerOption {
    return func(x *CollectorWorker) {
        x.batchSize = batchSize
    }
}

// WithCollectorWorkerFlushInterval 刷新间隔
func WithCollectorWorkerFlushInterval(flushInterval time.Duration) CollectorWorkerOption {
    return func(x *CollectorWorker) {
        x.flushInterval = flushInterval
    }
}
