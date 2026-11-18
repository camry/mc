package mc

// MapBaseIncr 递增映射的值（在写锁内完成读改写，并发安全）
func MapBaseIncr[K MapKeyType, V BaseTypeNumber](x *MapBase[K, V], key K, value V) V {
    x.mu.Lock()
    defer x.mu.Unlock()
    v := x.v[key]
    v += value
    x.v[key] = v
    x.Set(key, v)
    return v
}

// MapBaseDecr 递减映射的值（在写锁内完成读改写，并发安全）
func MapBaseDecr[K MapKeyType, V BaseTypeNumber](x *MapBase[K, V], key K, value V) V {
    x.mu.Lock()
    defer x.mu.Unlock()
    v := x.v[key]
    v -= value
    x.v[key] = v
    x.Set(key, v)
    return v
}

// MapMessageGetOrPut 查找指定键的映射对象，如果不存在则创建并返回（并发安全）
func MapMessageGetOrPut[K MapKeyType, V IMessage](x *MapMessage[K, V], key K) V {
    return x.GetOrPutNewMessage(key)
}
