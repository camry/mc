package mc

import (
    "errors"
    "reflect"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

type MapMessage[T comparable, V IMessage] struct {
    Map[T, V]
    creator func(parent IMessage, key any) V
}

func NewMapMessage[T comparable, V IMessage](parent IMessage, key string, creator func(parent IMessage, key any) V) *MapMessage[T, V] {
    return &MapMessage[T, V]{
        Map: Map[T, V]{
            parent: parent,
            key:    key,
            v:      make(map[T]V),
        },
        creator: creator,
    }
}

func (x *MapMessage[T, V]) MessageType() MessageType {
    return MessageTypeMapMessage
}

func (x *MapMessage[T, V]) Key() any {
    return x.key
}

func (x *MapMessage[T, V]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Set(x.parent, key, val, nestedMessages...)
}

func (x *MapMessage[T, V]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Unset(x.parent, key, nestedMessages...)
}

func (x *MapMessage[T, V]) Decode(vr bsonrw.ValueReader, _ any) error {
    if vr.Type() == bson.TypeEmbeddedDocument {
        doc, err := vr.ReadDocument()
        if err != nil {
            return err
        }
        for {
            ek, evr, err1 := doc.ReadElement()
            if errors.Is(err1, bsonrw.ErrEOD) {
                break
            }
            if err1 != nil {
                return err1
            }
            if evr.Type() == bson.TypeNull {
                return evr.ReadNull()
            }
            if evr.Type() == bson.TypeUndefined {
                return evr.ReadUndefined()
            }
            key, err1 := toMapKey[T](ek, reflect.TypeOf(x.v).Key().Kind())
            if err1 != nil {
                return err1
            }
            val := x.creator(x, key)
            decodeMessage, err1 := GetDecode(val)
            if err1 != nil {
                return err1
            }
            err1 = decodeMessage.Decode(evr, val)
            if err1 != nil {
                return err1
            }
            x.v[key] = val
        }
    }
    return nil
}

func (x *MapMessage[T, V]) PutNewMessage(key T) V {
    value := x.creator(x, key)
    x.v[key] = value
    if v, ok := any(value).(IMarshalBsonM); ok {
        x.Set(x.parent, key, v.Marshal())
    }
    return value
}

func (x *MapMessage[T, V]) Delete(key T) {
    delete(x.v, key)
    x.Unset(x.parent, key)
}

func (x *MapMessage[T, V]) Clear() {
    for k := range x.v {
        x.Unset(x.parent, k)
    }
    x.v = make(map[T]V)
}
