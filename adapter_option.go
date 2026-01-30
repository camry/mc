package mc

type AdapterOption[T IMessageTop, K BaseType] func(x *Adapter[T, K])

// WithCollector 设置适配器的收集器
func WithCollector[T IMessageTop, K BaseType](collector ICollector) AdapterOption[T, K] {
    return func(x *Adapter[T, K]) {
        x.collector = collector
    }
}

// WithCollectorWorkerOption 设置适配器的收集器工作选项
func WithCollectorWorkerOption[T IMessageTop, K BaseType](opts ...CollectorWorkerOption) AdapterOption[T, K] {
    return func(x *Adapter[T, K]) {
        x.collectorWorkerOption = opts
    }
}
