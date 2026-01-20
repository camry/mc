package mc

import (
    "errors"
    "reflect"

    "github.com/camry/g/gerrors/gerror"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

type MapBase[T comparable, V BaseType] struct {
    Map[T, V]
}

func NewMapBase[T comparable, V BaseType](parent IMessage, key string) *MapBase[T, V] {
    return &MapBase[T, V]{
        Map: Map[T, V]{
            parent: parent,
            key:    key,
            v:      make(map[T]V),
        },
    }
}

func (x *MapBase[T, V]) MessageType() MessageType {
    return MessageTypeMapBase
}

func (x *MapBase[T, V]) Key() any {
    return x.key
}

func (x *MapBase[T, V]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Set(x.parent, key, val, nestedMessages...)
}

func (x *MapBase[T, V]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Unset(x.parent, key, nestedMessages...)
}

func (x *MapBase[T, V]) Decode(vr bsonrw.ValueReader, _ any) error {
    if vr.Type() == bson.TypeEmbeddedDocument {
        doc, err := vr.ReadDocument()
        if err != nil {
            return err
        }
        for {
            k, evr, err1 := doc.ReadElement()
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
            var (
                key T
                val V
            )
            key, err1 = toMapKey[T](k, reflect.TypeOf(x.v).Key().Kind())
            if err1 != nil {
                return err1
            }
            kind := reflect.TypeOf(x.v).Elem().Kind()
            switch kind {
            case reflect.Int32:
                v, err2 := evr.ReadInt32()
                if err2 != nil {
                    return err2
                }
                val = any(v).(V)
            case reflect.Int64:
                v, err2 := evr.ReadInt64()
                if err2 != nil {
                    return err2
                }
                val = any(v).(V)
            case reflect.Float32:
                v, err2 := evr.ReadDouble()
                if err2 != nil {
                    return err2
                }
                val = any(float32(v)).(V)
            case reflect.Float64:
                v, err2 := evr.ReadDouble()
                if err2 != nil {
                    return err2
                }
                val = any(v).(V)
            case reflect.String:
                v, err2 := evr.ReadString()
                if err2 != nil {
                    return err2
                }
                val = any(v).(V)
            default:
                err2 := gerror.Newf("unknow BaseType %s", kind)
                if err2 != nil {
                    return err2
                }
            }
            x.v[key] = val
        }
    }
    return nil
}

func (x *MapBase[T, V]) Put(key T, value V) {
    x.v[key] = value
    x.Set(x.parent, key, value)
}

func (x *MapBase[T, V]) Delete(key T) {
    delete(x.v, key)
    x.Unset(x.parent, key)
}

func (x *MapBase[T, V]) Clear() {
    for k := range x.v {
        x.Unset(x.parent, k)
    }
    x.v = make(map[T]V)
}
