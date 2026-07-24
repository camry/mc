package mc

import (
    "reflect"
    "slices"

    "go.mongodb.org/mongo-driver/v2/bson"
)

// List 列表结构体，用于管理同类型元素的有序集合
// T 列表元素的类型
type List[T any] struct {
    parent IMessage // parent 父级消息对象
    key    string   // key 当前列表在父级中的键
    v      []T      // v 存储列表元素的实际切片
}

// Set 设置列表中某个元素的值
// key 要设置的字段键
// val 要设置的值
// nestedMessages 嵌套消息列表
func (x *List[T]) Set(key, val any, nestedMessages ...IMessage) {
    x.parent.Set(key, val, nestedMessages...) // 委托给父级设置
}

// Unset 取消设置列表中某个元素的值
// key 要取消设置的字段键
// nestedMessages 嵌套消息列表
func (x *List[T]) Unset(key any, nestedMessages ...IMessage) {
    x.parent.Unset(key, nestedMessages...) // 委托给父级取消设置
}

// Len 返回列表长度
func (x *List[T]) Len() int {
    return len(x.v)
}

// IsEmpty 检查列表是否为空
func (x *List[T]) IsEmpty() bool {
    return len(x.v) == 0
}

// Values 返回列表元素的切片
func (x *List[T]) Values() []T {
    return x.v
}

// Contains 检查列表是否包含指定元素
// item 要检查的元素
func (x *List[T]) Contains(item T) bool {
    for _, v := range x.v {
        if reflect.DeepEqual(v, item) { // 使用深度相等比较
            return true
        }
    }
    return false
}

// Find 在列表中查找满足条件的元素
// predicate 判断元素是否满足条件的函数
func (x *List[T]) Find(predicate func(item T) bool) (T, bool) {
    for _, v := range x.v {
        if predicate(v) {
            return v, true
        }
    }
    var zero T // 返回零值和 false
    return zero, false
}

// Filter 过滤出满足条件的元素
// predicate 判断元素是否满足条件的函数
func (x *List[T]) Filter(predicate func(item T) bool) []T {
    result := make([]T, 0, len(x.v)) // 预分配容量
    for _, v := range x.v {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

// ForEach 对列表中每个元素执行指定函数
// fn 接收索引和元素作为参数的函数
func (x *List[T]) ForEach(fn func(index int, item T)) {
    for i, v := range x.v {
        fn(i, v)
    }
}

// Append 向列表末尾添加元素
// item 要添加的元素
func (x *List[T]) Append(item T) {
    x.v = append(x.v, item)          // 添加元素到切片
    x.parent.Set(x.key, x.Marshal()) // 更新父级中的值
}

// Remove 移除指定索引位置的元素
// index 要移除元素的索引
func (x *List[T]) Remove(index int) {
    x.v = slices.Delete(x.v, index, index+1) // 删除指定索引的元素
    x.parent.Set(x.key, x.Marshal())         // 更新父级中的值
}

// Clear 清空列表中的所有元素
func (x *List[T]) Clear() {
    x.v = x.v[:0]                    // 保留底层数组，只清除元素
    x.parent.Set(x.key, x.Marshal()) // 更新父级中的值
}

// Marshal 将列表转换为bson.A格式
func (x *List[T]) Marshal() bson.A {
    a := bson.A{}
    if x.Len() > 0 {
        for _, v := range x.v {
            if vv, ok := any(v).(IMarshalBsonM); ok {
                a = append(a, vv.Marshal()) // 如果元素实现了 IMarshalBsonM 接口，则调用其 Marshal 方法
            } else {
                a = append(a, v) // 否则直接添加元素
            }
        }
    }
    return a
}
