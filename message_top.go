package mc

type MessageTop[T IMessageTop, K comparable] struct {
    a IAdapter[T, K]
}

func (m *MessageTop[T, K]) MessageType() MessageType {
    return MessageTypeTop
}

func (m *MessageTop[T, K]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    m.a.Set(m.a.GetMessageTop(), nestedMessages, key, val)
}

func (m *MessageTop[T, K]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    m.a.Unset(m.a.GetMessageTop(), nestedMessages, key)
}

func (m *MessageTop[T, K]) Bind(adapter IAdapter[T, K], t T, k K) {
    m.a = adapter
    m.a.Bind(t, k)
}

func (m *MessageTop[T, K]) Key() any {
    return m.a.GetKey()
}
