package mc

import (
    "errors"
    "reflect"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

// MapMessage 消息类型映射结构体，用于管理消息对象的映射
// T 键类型，必须是可比较的
// V 值类型，必须实现IMessage接口
type MapMessage[T MapKeyType, V IMessage] struct {
    Map[T, V]                                  // 继承 Map 的所有方法
    creator   func(parent IMessage, key any) V // 创建新消息对象的工厂函数
}

// NewMapMessage 创建新的消息类型映射实例
// parent 父级消息对象
// key 当前映射在父级中的键
// creator 创建新消息对象的工厂函数
func NewMapMessage[T MapKeyType, V IMessage](parent IMessage, key string, creator func(parent IMessage, key any) V) *MapMessage[T, V] {
    return &MapMessage[T, V]{
        Map: Map[T, V]{
            parent: parent,
            key:    key,
            v:      make(map[T]V), // 初始化映射
        },
        creator: creator, // 设置创建器函数
    }
}

// MessageType 返回消息类型
func (x *MapMessage[T, V]) MessageType() MessageType {
    return MessageTypeMapMessage
}

// Key 返回当前映射的键
func (x *MapMessage[T, V]) Key() any {
    return x.key
}

// Set 设置字段值，添加当前映射到嵌套消息链
// key 要设置的字段键
// val 要设置的值
// nestedMessages 嵌套消息列表
func (x *MapMessage[T, V]) Set(key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x) // 将当前映射添加到嵌套消息链
    x.parent.Set(key, val, nestedMessages...)  // 委托给父级设置
}

// Unset 取消设置字段值，添加当前映射到嵌套消息链
// key 要取消设置的字段键
// nestedMessages 嵌套消息列表
func (x *MapMessage[T, V]) Unset(key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x) // 将当前映射添加到嵌套消息链
    x.parent.Unset(key, nestedMessages...)     // 委托给父级取消设置
}

// Decode 从bson读取器中解码数据到当前映射
// vr bson值读取器
// _ 未使用的参数
func (x *MapMessage[T, V]) Decode(vr bsonrw.ValueReader, _ any) error {
    if vr.Type() == bson.TypeEmbeddedDocument {
        doc, err := vr.ReadDocument() // 开始读取文档
        if err != nil {
            return err
        }
        for {
            ek, evr, err1 := doc.ReadElement()  // 读取文档中的每个元素
            if errors.Is(err1, bsonrw.ErrEOD) { // 到达文档末尾
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
            // 转换键
            key, err1 := toMapKey[T](ek, reflect.TypeOf(x.v).Key().Kind())
            if err1 != nil {
                return err1
            }
            // 创建新消息对象
            val := x.creator(x, key)
            // 获取解码器
            decodeMessage, err1 := GetDecode(val)
            if err1 != nil {
                return err1
            }
            // 解码消息对象
            err1 = decodeMessage.Decode(evr, val)
            if err1 != nil {
                return err1
            }
            x.v[key] = val // 将键值对添加到映射
        }
    }
    return nil
}

// PutNewMessage 向映射中添加一个新的消息对象并返回该对象
// key 键
func (x *MapMessage[T, V]) PutNewMessage(key T) V {
    value := x.creator(x, key) // 创建新消息对象
    x.v[key] = value           // 添加到映射
    if v, ok := any(value).(IMarshalBsonM); ok {
        x.Set(key, v.Marshal()) // 如果实现了IMarshalBsonM接口，则设置序列化后的值
    }
    return value
}

// Delete 从映射中删除指定键的键值对
// key 要删除的键
func (x *MapMessage[T, V]) Delete(key T) {
    delete(x.v, key) // 从映射中删除键值对
    x.Unset(key)     // 从父级中取消设置
}
