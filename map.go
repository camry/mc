package mc

import (
    "github.com/spf13/cast"
    "go.mongodb.org/mongo-driver/bson"
)

type Map[T comparable, V any] struct {
    parent IMessage
    key    string
    v      map[T]V
}

func (x *Map[T, V]) Get(key T) (V, bool) {
    value, exists := x.v[key]
    return value, exists
}

func (x *Map[T, V]) Contains(key T) bool {
    _, exists := x.v[key]
    return exists
}

func (x *Map[T, V]) Size() int {
    return len(x.v)
}

func (x *Map[T, V]) IsEmpty() bool {
    return len(x.v) == 0
}

func (x *Map[T, V]) Keys() []T {
    keys := make([]T, 0, len(x.v))
    for k := range x.v {
        keys = append(keys, k)
    }
    return keys
}

func (x *Map[T, V]) Values() []V {
    values := make([]V, 0, len(x.v))
    for _, v := range x.v {
        values = append(values, v)
    }
    return values
}

func (x *Map[T, V]) GetAll() map[T]V {
    result := make(map[T]V, len(x.v))
    for k, v := range x.v {
        result[k] = v
    }
    return result
}

func (x *Map[T, V]) Marshal() bson.M {
    m := bson.M{}
    for k, v := range x.v {
        if vv, ok := any(v).(IMarshalBsonM); ok {
            m[cast.ToString(k)] = vv.Marshal()
        } else {
            m[cast.ToString(k)] = v
        }
    }
    return m
}
