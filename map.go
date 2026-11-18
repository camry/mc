package mc

import (
    "maps"
    "sync"

    "github.com/spf13/cast"
    "go.mongodb.org/mongo-driver/v2/bson"
)

// Map 映射结构体，用于管理键值对集合，可并发读写
// T 键类型，必须是可比较的
// V 值类型
type Map[T MapKeyType, V any] struct {
    mu     sync.RWMutex // mu 保护 v 的并发访问
    parent IMessage     // parent 父级消息对象
    key    string       // key 当前映射在父级中的键
    v      map[T]V      // v 存储键值对的实际映射
}

// Get 获取指定键对应的值
// key 要获取的键
func (x *Map[T, V]) Get(key T) (V, bool) {
    x.mu.RLock()
    defer x.mu.RUnlock()
    value, exists := x.v[key]
    return value, exists
}

// Contains 检查映射是否包含指定键
// key 要检查的键
func (x *Map[T, V]) Contains(key T) bool {
    x.mu.RLock()
    defer x.mu.RUnlock()
    _, exists := x.v[key]
    return exists
}

// Size 返回映射中键值对的数量
func (x *Map[T, V]) Size() int {
    x.mu.RLock()
    defer x.mu.RUnlock()
    return len(x.v)
}

// IsEmpty 检查映射是否为空
func (x *Map[T, V]) IsEmpty() bool {
    x.mu.RLock()
    defer x.mu.RUnlock()
    return len(x.v) == 0
}

// Keys 返回映射中所有键的切片
func (x *Map[T, V]) Keys() []T {
    x.mu.RLock()
    defer x.mu.RUnlock()
    keys := make([]T, 0, len(x.v)) // 预分配容量
    for k := range x.v {
        keys = append(keys, k)
    }
    return keys
}

// Values 返回映射中所有值的切片
func (x *Map[T, V]) Values() []V {
    x.mu.RLock()
    defer x.mu.RUnlock()
    values := make([]V, 0, len(x.v)) // 预分配容量
    for _, v := range x.v {
        values = append(values, v)
    }
    return values
}

// GetAll 返回映射的完整副本
func (x *Map[T, V]) GetAll() map[T]V {
    x.mu.RLock()
    defer x.mu.RUnlock()
    result := make(map[T]V, len(x.v)) // 预分配容量
    maps.Copy(result, x.v)
    return result
}

// Range 遍历映射中的每个键值对
// 遍历的是遍历开始时的快照，回调中可安全调用 Put/Delete 等写方法
func (x *Map[T, V]) Range(fn func(key T, value V) bool) {
    x.mu.RLock()
    snapshot := make(map[T]V, len(x.v)) // 拷贝快照，避免遍历期间持锁导致自死锁
    maps.Copy(snapshot, x.v)
    x.mu.RUnlock()
    for k, v := range snapshot {
        if !fn(k, v) {
            break
        }
    }
}

// Clear 清空映射中的所有键值对
func (x *Map[T, V]) Clear() {
    x.mu.Lock()
    defer x.mu.Unlock()
    x.v = make(map[T]V)                 // 创建一个新的空映射
    x.parent.Set(x.key, x.marshalLocked()) // 更新父级中的值
}

// Marshal 将映射转换为bson.M格式
func (x *Map[T, V]) Marshal() bson.M {
    x.mu.RLock()
    defer x.mu.RUnlock()
    return x.marshalLocked()
}

// marshalLocked 在已持有锁的前提下将映射转换为 bson.M
func (x *Map[T, V]) marshalLocked() bson.M {
    m := bson.M{}
    for k, v := range x.v {
        if vv, ok := any(v).(IMarshalBsonM); ok {
            m[cast.ToString(k)] = vv.Marshal() // 如果值实现了 IMarshalBsonM 接口，则调用其 Marshal 方法
        } else {
            m[cast.ToString(k)] = v // 否则直接使用值
        }
    }
    return m
}
