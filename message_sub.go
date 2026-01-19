package mc

type MessageSub struct {
    parent IMessage
    key    any
}

func (x *MessageSub) MessageType() MessageType {
    return MessageTypeSub
}

func (x *MessageSub) Key() any {
    return x.key
}

func (x *MessageSub) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Set(x.parent, key, val, nestedMessages...)
}

func (x *MessageSub) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Unset(x.parent, key, nestedMessages...)
}

func (x *MessageSub) Bind(parent IMessage, key any) {
    x.parent = parent
    x.key = key
}
