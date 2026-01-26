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
