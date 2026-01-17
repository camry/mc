package mc

import "go.mongodb.org/mongo-driver/bson/bsonrw"

type DecodeField[T any] struct {
    handler func(vr bsonrw.ValueReader, v T) error
}

func NewDecodeField[T any](handler func(vr bsonrw.ValueReader, v T) error) IDecode {
    return &DecodeField[T]{handler: handler}
}

func (x *DecodeField[T]) Decode(vr bsonrw.ValueReader, v any) error {
    return x.handler(vr, v.(T))
}
