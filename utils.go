package mc

import (
    "reflect"
    "strings"

    "github.com/camry/g/gerrors/gerror"
    "github.com/spf13/cast"
)

func GetStructName(i any) string {
    t := reflect.TypeOf(i)
    for t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    return t.Name()
}

func toNestedMessageFieldKey(nestedMessages []IMessage, key any) string {
    if len(nestedMessages) > 0 {
        var builder strings.Builder
        builder.WriteString(toNestedMessageKey(nestedMessages))
        builder.WriteString(".")
        builder.WriteString(cast.ToString(key))
        return builder.String()
    }
    return cast.ToString(key)
}

func toNestedMessageKey(nestedMessages []IMessage) string {
    if len(nestedMessages) == 0 {
        return ""
    }
    var builder strings.Builder
    for i := len(nestedMessages) - 1; i >= 0; i-- {
        message := nestedMessages[i]
        if builder.Len() > 0 {
            builder.WriteString(".")
        }
        builder.WriteString(cast.ToString(message.Key()))
    }
    return builder.String()
}

func toMapKey[T comparable](key string, kind reflect.Kind) (T, error) {
    var t T
    switch kind {
    case reflect.Int32:
        t = any(cast.ToInt32(key)).(T)
    case reflect.Int64:
        t = any(cast.ToInt64(key)).(T)
    case reflect.Float32:
        t = any(cast.ToFloat32(key)).(T)
    case reflect.Float64:
        t = any(cast.ToFloat64(key)).(T)
    case reflect.String:
        t = any(key).(T)
    default:
        return t, gerror.Newf("unsupported map key type: %s", kind)
    }
    return t, nil
}
