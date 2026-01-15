package mc

type MessageType int8

const (
    MessageTypeTop MessageType = iota
    MessageTypeSub
    MessageTypeList
    MessageTypeMapBase
    MessageTypeMapNested
)

type IAdapter[K comparable] interface {
    Collect(k K, nestedMessages []IMessage, key, val any)
}

type IMessage interface {
    MessageType() MessageType
    Key() any
    Collect(self IMessage, key, val any, nestedMessages ...IMessage)
    Delete(self IMessage, key any)
}

type IMessageTop interface {
    IMessage
}

type IMessageSub interface {
    IMessage

    ParentMessage() IMessage
}
