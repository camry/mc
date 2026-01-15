package mc

type MessageTop[K comparable] struct {
    a IAdapter[K]
    k K
}

func (m *MessageTop[K]) MessageType() MessageType {
    return MessageTypeTop
}

func (m *MessageTop[K]) Collect(self IMessage, key, val any, nestedMessages ...IMessage) {
    m.a.Collect(m.k, nestedMessages, key, val)
}

func (m *MessageTop[K]) Delete(self IMessage, key any) {

}

func (m *MessageTop[K]) Bind(adapter IAdapter[K], k K) {
    m.a = adapter
    m.k = k
}

func (m *MessageTop[K]) Key() any {
    return m.k
}
