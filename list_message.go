package mc

import (
    "errors"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

// ListMessage 消息类型列表结构体，用于管理消息对象的列表
// T 消息类型，必须实现IMessage接口
type ListMessage[T IMessage] struct {
    List[T]                                  // 继承 List 的所有方法
    creator func(parent IMessage, key any) T // 创建新消息对象的工厂函数
}

// NewListMessage 创建新的消息类型列表实例
// parent 父级消息对象
// key 当前列表在父级中的键
// creator 创建新消息对象的工厂函数
func NewListMessage[T IMessage](parent IMessage, key string, creator func(parent IMessage, key any) T) *ListMessage[T] {
    return &ListMessage[T]{
        List: List[T]{
            parent: parent,
            key:    key,
        },
        creator: creator, // 设置创建器函数
    }
}

// MessageType 返回消息类型
func (x *ListMessage[T]) MessageType() MessageType {
    return MessageTypeListMessage
}

// Key 返回当前列表的键
func (x *ListMessage[T]) Key() any {
    return x.key
}

// Set 设置字段值，添加当前列表到嵌套消息链
// self 当前列表对象
// key 要设置的字段键
// val 要设置的值
// nestedMessages 嵌套消息列表
func (x *ListMessage[T]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)        // 将当前列表添加到嵌套消息链
    x.List.Set(x.parent, key, val, nestedMessages...) // 委托给父类设置
}

// Unset 取消设置字段值，添加当前列表到嵌套消息链
// self 当前列表对象
// key 要取消设置的字段键
// nestedMessages 嵌套消息列表
func (x *ListMessage[T]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)     // 将当前列表添加到嵌套消息链
    x.List.Unset(x.parent, key, nestedMessages...) // 委托给父类取消设置
}

// Decode 从bson读取器中解码数据到当前列表
// vr bson值读取器
// _ 未使用的参数
func (x *ListMessage[T]) Decode(vr bsonrw.ValueReader, _ any) error {
    if vr.Type() == bson.TypeArray {
        arr, err := vr.ReadArray() // 开始读取数组
        if err != nil {
            return err
        }
        for {
            avr, err1 := arr.ReadValue()        // 读取数组中的每个值
            if errors.Is(err1, bsonrw.ErrEOA) { // 到达数组末尾
                break
            }
            if err1 != nil {
                return err1
            }
            if avr.Type() != bson.TypeEmbeddedDocument { // 检查是否为嵌入文档类型
                continue
            }
            // 创建新消息对象
            val := x.creator(x, x.Len()) // 使用当前长度作为键
            // 获取解码器
            decodeMessage, err1 := GetDecode(val)
            if err1 != nil {
                return err1
            }
            // 解码消息对象
            err1 = decodeMessage.Decode(avr, val)
            if err1 != nil {
                return err1
            }
            x.v = append(x.v, val) // 添加解码后的消息对象到列表
        }
    }
    return nil
}

// AppendNewMessage 向列表追加一个新的消息对象并返回该对象
func (x *ListMessage[T]) AppendNewMessage() T {
    item := x.creator(x, x.Len()) // 使用当前长度作为键创建新消息对象
    x.Append(item)                // 添加到列表
    return item                   // 返回新创建的对象
}
