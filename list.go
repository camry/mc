package mc

import (
    "reflect"
    "slices"

    "go.mongodb.org/mongo-driver/bson"
)

type List[T any] struct {
    parent IMessage
    key    string
    v      []T
}

func (x *List[T]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    x.parent.Set(x.parent, key, val, nestedMessages...)
}

func (x *List[T]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    x.parent.Unset(x.parent, key, nestedMessages...)
}

func (x *List[T]) Len() int {
    return len(x.v)
}

func (x *List[T]) IsEmpty() bool {
    return len(x.v) == 0
}

func (x *List[T]) Contains(item T) bool {
    for _, v := range x.v {
        if reflect.DeepEqual(v, item) {
            return true
        }
    }
    return false
}

func (x *List[T]) Find(predicate func(item T) bool) (T, bool) {
    for _, v := range x.v {
        if predicate(v) {
            return v, true
        }
    }
    var zero T
    return zero, false
}

func (x *List[T]) Filter(predicate func(item T) bool) []T {
    result := make([]T, 0, len(x.v))
    for _, v := range x.v {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

func (x *List[T]) ForEach(fn func(index int, item T)) {
    for i, v := range x.v {
        fn(i, v)
    }
}

func (x *List[T]) Append(item T) {
    x.v = append(x.v, item)
    x.parent.Set(x.parent, x.key, x.Marshal())
}

func (x *List[T]) Remove(index int) {
    x.v = slices.Delete(x.v, index, index+1)
    x.parent.Set(x.parent, x.key, x.Marshal())
}

func (x *List[T]) Marshal() bson.A {
    a := bson.A{}
    for _, v := range x.v {
        if vv, ok := any(v).(IMarshalBsonM); ok {
            a = append(a, vv.Marshal())
        } else {
            a = append(a, v)
        }
    }
    return a
}
