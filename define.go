package mc

import (
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
    "golang.org/x/exp/constraints"
)

// MessageType 表示消息类型的枚举值
type MessageType int8

const (
    MessageTypeTop         MessageType = iota // MessageTypeTop 顶级消息类型
    MessageTypeSub                            // MessageTypeSub 子级消息类型
    MessageTypeListBase                       // MessageTypeListBase 基础类型列表消息类型
    MessageTypeListMessage                    // MessageTypeListMessage 消息类型列表消息类型
    MessageTypeMapBase                        // MessageTypeMapBase 基础类型映射消息类型
    MessageTypeMapMessage                     // MessageTypeMapMessage 消息类型映射消息类型
)

// BaseType 定义基础类型约束，包括整型、浮点型和字符串
type BaseType interface {
    BaseTypeInteger | ~string
}

// BaseTypeInteger 定义基础类型整型约束，包括int32、int64和浮点型
type BaseTypeInteger interface {
    ~int32 | ~int64 | constraints.Float
}

// IMarshalBsonM 定义BSON映射接口，用于将对象转换为bson.M格式
type IMarshalBsonM interface {
    Marshal() bson.M
}

// IAdapter 定义适配器接口，用于管理消息对象和键值对
type IAdapter[T IMessageTop, K comparable] interface {
    Bind(t T, k K)                                                       // Bind 绑定消息对象和键
    GetMessageTop() T                                                    // GetMessageTop 获取顶级消息对象
    GetKey() K                                                           // GetKey 获取键值
    Set(topMessage IMessageTop, nestedMessages []IMessage, key, val any) // Set 设置指定字段的值
    Unset(topMessage IMessageTop, nestedMessages []IMessage, key any)    // Unset 取消设置指定字段的值
}

// IMessage 定义消息接口，所有消息类型都应实现该接口
type IMessage interface {
    MessageType() MessageType                                    // MessageType 返回消息类型
    Key() any                                                    // Key 返回消息的键
    Set(self IMessage, key, val any, nestedMessages ...IMessage) // Set 设置字段值，支持嵌套消息
    Unset(self IMessage, key any, nestedMessages ...IMessage)    // Unset 取消设置字段值，支持嵌套消息
}

// IMessageTop 定义顶级消息接口，扩展自IMessage和IMarshalBsonM
type IMessageTop interface {
    IMessage
    IMarshalBsonM

    Table() string // Table 返回对应的数据表名
}

// IMessageSub 定义子级消息接口，扩展自IMessage和IMarshalBsonM
type IMessageSub interface {
    IMessage
    IMarshalBsonM
}

// IDecode 定义解码接口，用于从bson读取器中解析数据
type IDecode interface {
    Decode(vr bsonrw.ValueReader, v any) error // Decode 从bson读取器中解码数据到目标对象
}
