package mc

import (
    "errors"
    "slices"

    "go.mongodb.org/mongo-driver/v2/bson"
)

// ListBase 基础类型列表结构体，继承自 List
// T 基础类型，必须实现BaseType约束
type ListBase[T BaseType] struct {
    List[T] // 继承 List 的所有方法
}

// NewListBase 创建新的基础类型列表实例
// parent 父级消息对象
// key 当前列表在父级中的键
func NewListBase[T BaseType](parent IMessage, key string) *ListBase[T] {
    return &ListBase[T]{
        List: List[T]{
            parent: parent,
            key:    key,
        },
    }
}

// MessageType 返回消息类型
func (x *ListBase[T]) MessageType() MessageType {
    return MessageTypeListBase
}

// Key 返回当前列表的键
func (x *ListBase[T]) Key() any {
    return x.key
}

// Set 设置字段值，添加当前列表到嵌套消息链
// key 要设置的字段键
// val 要设置的值
// nestedMessages 嵌套消息列表
func (x *ListBase[T]) Set(key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x) // 将当前列表添加到嵌套消息链
    x.List.Set(key, val, nestedMessages...)    // 委托给父类设置
}

// Unset 取消设置字段值，添加当前列表到嵌套消息链
// key 要取消设置的字段键
// nestedMessages 嵌套消息列表
func (x *ListBase[T]) Unset(key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x) // 将当前列表添加到嵌套消息链
    x.List.Unset(key, nestedMessages...)       // 委托给父类取消设置
}

// Decode 从bson读取器中解码数据到当前列表
// vr bson值读取器
// _ 未使用的参数
func (x *ListBase[T]) Decode(vr bson.ValueReader, _ any) error {
    if vr.Type() != bson.TypeArray {
        return vr.Skip() // 非数组类型（含 null）直接跳过
    }
    arr, err := vr.ReadArray() // 开始读取数组
    if err != nil {
        return err
    }
    for {
        avr, err1 := arr.ReadValue()      // 读取数组中的每个值
        if errors.Is(err1, bson.ErrEOA) { // 到达数组末尾
            break
        }
        if err1 != nil {
            return err1
        }
        // 跳过无法解码为目标类型的值
        if !canBaseDecode[T](avr.Type()) {
            if err1 = avr.Skip(); err1 != nil {
                return err1
            }
            continue
        }
        val, err1 := BaseDecode[T](avr)
        if err1 != nil {
            return err1
        }
        x.v = append(x.v, val) // 添加解码后的值到列表
    }
    return nil
}

// Remove 根据值移除列表中的元素
// value 要移除的元素
func (x *ListBase[T]) Remove(value T) {
    index := slices.Index(x.v, value) // 查找元素索引
    if index == -1 {
        return // 元素不存在，直接返回
    }
    x.List.Remove(index) // 调用父类的 Remove 方法
}
