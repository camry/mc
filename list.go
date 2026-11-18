package mc

import (
    "reflect"
    "slices"
    "sync"

    "go.mongodb.org/mongo-driver/v2/bson"
)

// List 列表结构体，用于管理同类型元素的有序集合，可并发读写
// T 列表元素的类型
type List[T any] struct {
    mu     sync.RWMutex // mu 保护 v 的并发访问
    parent IMessage     // parent 父级消息对象
    key    string       // key 当前列表在父级中的键
    v      []T          // v 存储列表元素的实际切片
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
    x.mu.RLock()
    defer x.mu.RUnlock()
    return len(x.v)
}

// IsEmpty 检查列表是否为空
func (x *List[T]) IsEmpty() bool {
    x.mu.RLock()
    defer x.mu.RUnlock()
    return len(x.v) == 0
}

// Values 返回列表元素的副本切片
func (x *List[T]) Values() []T {
    x.mu.RLock()
    defer x.mu.RUnlock()
    return slices.Clone(x.v)
}

// Contains 检查列表是否包含指定元素
// value 要检查的元素
func (x *List[T]) Contains(value T) bool {
    x.mu.RLock()
    defer x.mu.RUnlock()
    for _, v := range x.v {
        if reflect.DeepEqual(v, value) { // 使用深度相等比较
            return true
        }
    }
    return false
}

// Find 在列表中查找满足条件的元素
// predicate 判断元素是否满足条件的函数
func (x *List[T]) Find(predicate func(value T) bool) (T, bool) {
    x.mu.RLock()
    snapshot := slices.Clone(x.v) // 拷贝快照，回调中可安全写列表
    x.mu.RUnlock()
    for _, v := range snapshot {
        if predicate(v) {
            return v, true
        }
    }
    var zero T // 返回零值和 false
    return zero, false
}

// Filter 过滤出满足条件的元素
// predicate 判断元素是否满足条件的函数
func (x *List[T]) Filter(predicate func(value T) bool) []T {
    x.mu.RLock()
    snapshot := slices.Clone(x.v) // 拷贝快照，回调中可安全写列表
    x.mu.RUnlock()
    result := make([]T, 0, len(snapshot)) // 预分配容量
    for _, v := range snapshot {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

// Range 对列表中每个元素执行指定函数
// 遍历的是遍历开始时的快照，回调中可安全调用 Append/Remove 等写方法
// fn 接收索引和元素作为参数的函数
func (x *List[T]) Range(fn func(index int, value T) bool) {
    x.mu.RLock()
    snapshot := slices.Clone(x.v) // 拷贝快照，避免遍历期间持锁导致自死锁
    x.mu.RUnlock()
    for i, v := range snapshot {
        if !fn(i, v) {
            break
        }
    }
}

// Append 向列表末尾添加元素
// value 要添加的元素
func (x *List[T]) Append(value T) {
    x.mu.Lock()
    defer x.mu.Unlock()
    x.appendLocked(value)
}

// Remove 移除指定索引位置的元素
// index 要移除元素的索引
func (x *List[T]) Remove(index int) {
    x.mu.Lock()
    defer x.mu.Unlock()
    x.removeLocked(index)
}

// Clear 清空列表中的所有元素
func (x *List[T]) Clear() {
    x.mu.Lock()
    defer x.mu.Unlock()
    x.v = x.v[:0]                       // 保留底层数组，只清除元素
    x.parent.Set(x.key, x.marshalLocked()) // 更新父级中的值
}

// Marshal 将列表转换为bson.A格式
func (x *List[T]) Marshal() bson.A {
    x.mu.RLock()
    defer x.mu.RUnlock()
    return x.marshalLocked()
}

// appendLocked 在已持有写锁的前提下追加元素并同步父级
func (x *List[T]) appendLocked(value T) {
    x.v = append(x.v, value)            // 添加元素到切片
    x.parent.Set(x.key, x.marshalLocked()) // 更新父级中的值
}

// removeLocked 在已持有写锁的前提下删除指定索引的元素并同步父级
func (x *List[T]) removeLocked(index int) {
    x.v = slices.Delete(x.v, index, index+1) // 删除指定索引的元素
    x.parent.Set(x.key, x.marshalLocked())   // 更新父级中的值
}

// marshalLocked 在已持有锁的前提下将列表转换为 bson.A
func (x *List[T]) marshalLocked() bson.A {
    a := bson.A{}
    for _, v := range x.v {
        if vv, ok := any(v).(IMarshalBsonM); ok {
            a = append(a, vv.Marshal()) // 如果元素实现了 IMarshalBsonM 接口，则调用其 Marshal 方法
        } else {
            a = append(a, v) // 否则直接添加元素
        }
    }
    return a
}
