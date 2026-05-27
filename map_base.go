package mc

import (
    "errors"
    "reflect"

    "github.com/camry/g/v2/gerrors/gerror"
    "go.mongodb.org/mongo-driver/v2/bson"
)

// MapBase 基础类型映射结构体，继承自 Map
// T 键类型，必须是可比较的
// V 值类型，必须是基础类型
type MapBase[T MapKeyType, V BaseType] struct {
    Map[T, V] // 继承 Map 的所有方法
}

// NewMapBase 创建新的基础类型映射实例
// parent 父级消息对象
// key 当前映射在父级中的键
func NewMapBase[T MapKeyType, V BaseType](parent IMessage, key string) *MapBase[T, V] {
    return &MapBase[T, V]{
        Map: Map[T, V]{
            parent: parent,
            key:    key,
            v:      make(map[T]V), // 初始化映射
        },
    }
}

// MessageType 返回消息类型
func (x *MapBase[T, V]) MessageType() MessageType {
    return MessageTypeMapBase
}

// Key 返回当前映射的键
func (x *MapBase[T, V]) Key() any {
    return x.key
}

// Set 设置字段值，添加当前映射到嵌套消息链
// key 要设置的字段键
// val 要设置的值
// nestedMessages 嵌套消息列表
func (x *MapBase[T, V]) Set(key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x) // 将当前映射添加到嵌套消息链
    x.parent.Set(key, val, nestedMessages...)  // 委托给父级设置
}

// Unset 取消设置字段值，添加当前映射到嵌套消息链
// key 要取消设置的字段键
// nestedMessages 嵌套消息列表
func (x *MapBase[T, V]) Unset(key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x) // 将当前映射添加到嵌套消息链
    x.parent.Unset(key, nestedMessages...)     // 委托给父级取消设置
}

// Decode 从bson读取器中解码数据到当前映射
// vr bson值读取器
// _ 未使用的参数
func (x *MapBase[T, V]) Decode(vr bson.ValueReader, _ any) error {
    if vr.Type() == bson.TypeEmbeddedDocument {
        doc, err := vr.ReadDocument() // 开始读取文档
        if err != nil {
            return err
        }
        for {
            k, evr, err1 := doc.ReadElement() // 读取文档中的每个元素
            if errors.Is(err1, bson.ErrEOD) { // 到达文档末尾
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
            var (
                key  T
                val  any
                err2 error
            )
            // 转换键
            key, err1 = toMapKey[T](k, reflect.TypeOf(x.v).Key().Kind())
            if err1 != nil {
                return err1
            }
            // 获取值的目标类型
            valueType := reflect.TypeOf(x.v).Elem()
            // 根据值类型进行转换
            kind := valueType.Kind()
            switch kind {
            case reflect.Int32:
                val, err2 = evr.ReadInt32()
            case reflect.Int64:
                val, err2 = evr.ReadInt64()
            case reflect.Float32:
                vv, err3 := evr.ReadDouble()
                if err3 != nil {
                    return err3
                }
                val = float32(vv)
            case reflect.Float64:
                val, err2 = evr.ReadDouble()
            case reflect.String:
                val, err2 = evr.ReadString()
            default:
                return gerror.Newf("unknow BaseType %s", kind)
            }
            if err2 != nil {
                return err2
            }
            x.v[key] = reflect.ValueOf(val).Convert(valueType).Interface().(V) // 将键值对添加到映射
        }
    }
    return nil
}

// Put 向映射中添加或更新键值对
// key 键
// value 值
func (x *MapBase[T, V]) Put(key T, value V) {
    x.v[key] = value  // 更新映射中的值
    x.Set(key, value) // 更新父级中的值
}

// Delete 从映射中删除指定键的键值对
// key 要删除的键
func (x *MapBase[T, V]) Delete(key T) {
    delete(x.v, key) // 从映射中删除键值对
    x.Unset(key)     // 从父级中取消设置
}
