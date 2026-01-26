package mc

// MapBaseIncr 递增映射的值
func MapBaseIncr[K comparable, V BaseTypeInteger](x *MapBase[K, V], key K, value V) V {
    v, _ := x.Get(key)
    x.Put(key, v+value)
    return v
}

// MapBaseDecr 递减映射的值
func MapBaseDecr[K comparable, V BaseTypeInteger](x *MapBase[K, V], key K, value V) V {
    v, _ := x.Get(key)
    x.Put(key, v-value)
    return v
}

// MapMessageGetOrPut 查找指定键的映射对象，如果不存在则创建并返回
func MapMessageGetOrPut[K comparable, V IMessage](x *MapMessage[K, V], key K) V {
    v, ok := x.Get(key)
    if ok {
        return v
    }
    return x.PutNewMessage(key)
}
