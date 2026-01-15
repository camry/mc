package mc

import "github.com/camry/g/glog"

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

func (m *MessageSub) Collect(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, m)
    m.ParentMessage().Collect(m.parent, key, val, nestedMessages...)
}

func (m *MessageSub) Delete(self IMessage, key any) {
    glog.Fatalf("MessageSub Delete Not Impl")
}

func (m *MessageSub) ParentMessage() IMessage {
    return m.parent
}

func (m *MessageSub) Bind(parent IMessage, key any) {
    m.parent = parent
    m.key = key
}
