package mc

type MessageTop[T IMessageTop, K comparable] struct {
    a IAdapter[T, K]
}

func (x *MessageTop[T, K]) MessageType() MessageType {
    return MessageTypeTop
}

func (x *MessageTop[T, K]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    x.a.Set(x.a.GetMessageTop(), nestedMessages, key, val)
}

func (x *MessageTop[T, K]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    x.a.Unset(x.a.GetMessageTop(), nestedMessages, key)
}

func (x *MessageTop[T, K]) Bind(adapter IAdapter[T, K], t T, k K) {
    x.a = adapter
    x.a.Bind(t, k)
}

func (x *MessageTop[T, K]) Key() any {
    return x.a.GetKey()
}
