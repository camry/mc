package mc

import (
    "errors"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

type ListMessage[T IMessage] struct {
    List[T]
    creator func(parent IMessage, key any) T
}

func NewListMessage[T IMessage](parent IMessage, key string, creator func(parent IMessage, key any) T) *ListMessage[T] {
    return &ListMessage[T]{
        List: List[T]{
            parent: parent,
            key:    key,
        },
        creator: creator,
    }
}

func (x *ListMessage[T]) MessageType() MessageType {
    return MessageTypeListMessage
}

func (x *ListMessage[T]) Key() any {
    return x.key
}

func (x *ListMessage[T]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
}

func (x *ListMessage[T]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
}

func (x *ListMessage[T]) NewMessage() T {
    return x.creator(x, nil)
}

func (x *ListMessage[T]) Decode(vr bsonrw.ValueReader, _ any) error {
    if vr.Type() == bson.TypeArray {
        arr, err := vr.ReadArray()
        if err != nil {
            return err
        }
        for {
            avr, err1 := arr.ReadValue()
            if errors.Is(err1, bsonrw.ErrEOA) {
                break
            }
            if err1 != nil {
                return err1
            }
            if avr.Type() != bson.TypeEmbeddedDocument {
                continue
            }
            val := x.creator(x, nil)
            decodeMessage, err1 := GetDecode(val)
            if err1 != nil {
                return err1
            }
            err1 = decodeMessage.Decode(avr, val)
            if err1 != nil {
                return err1
            }
            x.v = append(x.v, val)
        }
    }
    return nil
}
