package mc

import "time"

// CollectorWorkerOption 收集器工作选项
type CollectorWorkerOption func(x *CollectorWorker)

// WithCollectorWorkerBufferSize 缓冲区大小
func WithCollectorWorkerBufferSize(bufferSize int) CollectorWorkerOption {
	return func(x *CollectorWorker) { x.bufferSize = bufferSize }
}

// WithCollectorWorkerBatchSize 批量大小
func WithCollectorWorkerBatchSize(batchSize int) CollectorWorkerOption {
	return func(x *CollectorWorker) { x.batchSize = batchSize }
}

// WithCollectorWorkerFlushInterval 刷新间隔
func WithCollectorWorkerFlushInterval(flushInterval time.Duration) CollectorWorkerOption {
	return func(x *CollectorWorker) { x.flushInterval = flushInterval }
}

// WithCollectorWorkerWriteTimeout 单次 BulkWrite 超时时间，<=0 表示不限制
func WithCollectorWorkerWriteTimeout(writeTimeout time.Duration) CollectorWorkerOption {
	return func(x *CollectorWorker) { x.writeTimeout = writeTimeout }
}

// WithCollectorWorkerMaxRetries BulkWrite 失败后的最大重试次数（总尝试次数 = maxRetries+1），<=0 表示不重试
func WithCollectorWorkerMaxRetries(maxRetries int) CollectorWorkerOption {
	return func(x *CollectorWorker) { x.maxRetries = maxRetries }
}

// WithCollectorWorkerRetryBaseDelay 重试退避基数（指数退避的起始值，每次重试翻倍）
func WithCollectorWorkerRetryBaseDelay(retryBaseDelay time.Duration) CollectorWorkerOption {
	return func(x *CollectorWorker) { x.retryBaseDelay = retryBaseDelay }
}

// WithCollectorWorkerRetryMaxDelay 重试退避上限
func WithCollectorWorkerRetryMaxDelay(retryMaxDelay time.Duration) CollectorWorkerOption {
	return func(x *CollectorWorker) { x.retryMaxDelay = retryMaxDelay }
}
