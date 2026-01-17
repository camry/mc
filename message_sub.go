package mc

type MessageSub struct {
    parent IMessage
    key    any
}

func (m *MessageSub) MessageType() MessageType {
    return MessageTypeSub
}

func (m *MessageSub) Key() any {
    return m.key
}

func (m *MessageSub) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, m)
    m.parent.Set(m.parent, key, val, nestedMessages...)
}

func (m *MessageSub) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, m)
    m.parent.Unset(m.parent, key, nestedMessages...)
}

func (m *MessageSub) Bind(parent IMessage, key any) {
    m.parent = parent
    m.key = key
}
