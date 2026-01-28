package mc

type AdapterOption[T IMessageTop, K comparable] func(x *Adapter[T, K])

// WithAdapterEnableSave 设置适配器是否保存数据
func WithAdapterEnableSave[T IMessageTop, K comparable](enableSave bool) AdapterOption[T, K] {
    return func(x *Adapter[T, K]) {
        x.enableSave = enableSave
    }
}

// WithAdapterSize 设置适配器的缓冲容量
func WithAdapterSize[T IMessageTop, K comparable](size int) AdapterOption[T, K] {
    return func(x *Adapter[T, K]) {
        x.size = size
    }
}
