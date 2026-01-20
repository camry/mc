package mc

import "go.mongodb.org/mongo-driver/bson/bsonrw"

// DecodeField 字段解码器结构体
// T 解码的目标类型
type DecodeField[T any] struct {
    handler func(vr bsonrw.ValueReader, v T) error // handler 解码处理器函数
}

// NewDecodeField 创建新的字段解码器
// handler 解码处理器函数
func NewDecodeField[T any](handler func(vr bsonrw.ValueReader, v T) error) IDecode {
    return &DecodeField[T]{handler: handler}
}

// Decode 执行解码操作
// vr bson值读取器
// v 解码目标对象
func (x *DecodeField[T]) Decode(vr bsonrw.ValueReader, v any) error {
    return x.handler(vr, v.(T)) // 调用处理器函数执行解码
}
