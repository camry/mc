package mc

import (
    "bytes"
    "reflect"
    "strings"

    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/spf13/cast"
    "go.mongodb.org/mongo-driver/v2/bson"
)

// GetStructName 获取结构体的名称（不包括包名）
// i 任意对象
func GetStructName(i any) string {
    t := reflect.TypeOf(i)        // 获取类型信息
    for t.Kind() == reflect.Ptr { // 如果是指针类型，获取指向的类型
        t = t.Elem()
    }
    return t.Name() // 返回类型名称
}

// toNestedMessageFieldKey 构建嵌套消息字段键
// nestedMessages 嵌套消息列表
// key 当前字段键
func toNestedMessageFieldKey(nestedMessages []IMessage, key any) string {
    if len(nestedMessages) > 0 {
        var builder strings.Builder
        builder.WriteString(toNestedMessageKey(nestedMessages)) // 添加嵌套消息键
        builder.WriteString(".")                                // 添加分隔符
        builder.WriteString(cast.ToString(key))                 // 添加当前键
        return builder.String()
    }
    return cast.ToString(key) // 如果没有嵌套消息，直接返回当前键
}

// toNestedMessageKey 构建嵌套消息键
// nestedMessages 嵌套消息列表
func toNestedMessageKey(nestedMessages []IMessage) string {
    if len(nestedMessages) == 0 {
        return ""
    }
    var builder strings.Builder
    for i := len(nestedMessages) - 1; i >= 0; i-- { // 从后往前遍历
        message := nestedMessages[i]
        if builder.Len() > 0 {
            builder.WriteString(".") // 添加分隔符
        }
        builder.WriteString(cast.ToString(message.Key())) // 添加消息键
    }
    return builder.String()
}

// toMapKey 将字符串键转换为指定类型的键
// key 字符串形式的键
// kind 目标类型种类
func toMapKey[T MapKeyType](key string, kind reflect.Kind) (T, error) {
    var k any
    // 获取目标类型
    targetType := reflect.TypeOf((*T)(nil)).Elem()
    switch kind {
    case reflect.Int32:
        k = cast.ToInt32(key)
    case reflect.Int64:
        k = cast.ToInt64(key)
    case reflect.Uint32:
        k = cast.ToUint32(key)
    case reflect.Uint64:
        k = cast.ToUint64(key)
    case reflect.String:
        k = key
    default:
        var t T
        return t, gerror.Newf("unsupported map key type: %s", kind)
    }
    return reflect.ValueOf(k).Convert(targetType).Interface().(T), nil
}

// MToValueReader 将 bson.M 对象转换为 bson.ValueReader 对象
// m bson.M 对象
// 返回 bson.ValueReader 对象
func MToValueReader(m bson.M) (bson.ValueReader, error) {
    raw, err := bson.Marshal(m)
    if err != nil {
        return nil, err
    }
    bsonRaw := bson.Raw(raw)
    return bson.NewDocumentReader(bytes.NewReader(bsonRaw)), nil
}
