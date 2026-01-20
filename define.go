package mc

import (
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
    "golang.org/x/exp/constraints"
)

type MessageType int8

const (
    MessageTypeTop MessageType = iota
    MessageTypeSub
    MessageTypeListBase
    MessageTypeListMessage
    MessageTypeMapBase
    MessageTypeMapMessage
)

type BaseType interface {
    ~int32 | ~int64 | constraints.Float | ~string
}

type IMarshalBsonM interface {
    Marshal() bson.M
}

type IAdapter[T IMessageTop, K comparable] interface {
    Bind(t T, k K)
    GetMessageTop() T
    GetKey() K
    Set(topMessage IMessageTop, nestedMessages []IMessage, key, val any)
    Unset(topMessage IMessageTop, nestedMessages []IMessage, key any)
}

type IMessage interface {
    MessageType() MessageType
    Key() any
    Set(self IMessage, key, val any, nestedMessages ...IMessage)
    Unset(self IMessage, key any, nestedMessages ...IMessage)
}

type IMessageTop interface {
    IMessage
    IMarshalBsonM

    Table() string
}

type IMessageSub interface {
    IMessage
    IMarshalBsonM
}

type IDecode interface {
    Decode(vr bsonrw.ValueReader, v any) error
}
