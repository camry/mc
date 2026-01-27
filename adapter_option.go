package mc

type AdapterOption[T IMessageTop, K comparable] func(x *Adapter[T, K])

func WithAdapterEnableSave[T IMessageTop, K comparable](enableSave bool) AdapterOption[T, K] {
    return func(x *Adapter[T, K]) {
        x.enableSave = enableSave
    }
}

func WithAdapterSize[T IMessageTop, K comparable](size int) AdapterOption[T, K] {
    return func(x *Adapter[T, K]) {
        x.size = size
    }
}
