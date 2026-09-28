package mc

import (
    "encoding/base64"
    "reflect"

    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/spf13/cast"
    "go.mongodb.org/mongo-driver/v2/bson"
)

// canBaseDecode 判断bson类型能否解码为基础类型T
// t bson类型
func canBaseDecode[T BaseType](t bson.Type) bool {
    switch t {
    case bson.TypeInt32, bson.TypeInt64, bson.TypeDouble, bson.TypeString:
        return true
    default:
        return false
    }
}

// BaseDecode 从bson值读取器中解码基础类型值，数值类型之间宽松转换
// vr bson值读取器
// T 基础类型
func BaseDecode[T BaseType](vr bson.ValueReader) (T, error) {
    var (
        zero T
        val  any
        err  error
    )
    valueType := reflect.TypeFor[T]()
    switch kind := valueType.Kind(); kind {
    case reflect.Int32:
        val, err = Int32Decode(vr)
    case reflect.Int64:
        val, err = Int64Decode(vr)
    case reflect.Uint32:
        val, err = UInt32Decode(vr)
    case reflect.Uint64:
        val, err = UInt64Decode(vr)
    case reflect.Float32:
        val, err = Float32Decode(vr)
    case reflect.Float64:
        val, err = Float64Decode(vr)
    case reflect.Bool:
        val, err = BoolDecode(vr)
    case reflect.String:
        val, err = StringDecode(vr)
    default:
        return zero, gerror.Newf("unsupported BaseType %s", kind)
    }
    if err != nil {
        return zero, err
    }
    return reflect.ValueOf(val).Convert(valueType).Interface().(T), nil // 转换为目标类型（支持枚举等派生类型）
}

// Float32Decode 从bson值读取器中解码float32值
// vr bson值读取器
func Float32Decode(vr bson.ValueReader) (float32, error) {
    v, err := Float64Decode(vr) // 先解码为float64再转换
    return cast.ToFloat32(v), err
}

// Float64Decode 从bson值读取器中解码float64值
// vr bson值读取器
func Float64Decode(vr bson.ValueReader) (float64, error) {
    switch vr.Type() {
    case bson.TypeInt32:
        v, err := vr.ReadInt32()
        return cast.ToFloat64(v), err
    case bson.TypeInt64:
        v, err := vr.ReadInt64()
        return cast.ToFloat64(v), err
    case bson.TypeBoolean:
        v, err := vr.ReadBoolean()
        return cast.ToFloat64(v), err
    case bson.TypeDouble:
        return vr.ReadDouble()
    case bson.TypeNull:
        return 0, vr.ReadNull()
    case bson.TypeUndefined:
        return 0, vr.ReadUndefined()
    default:
        return 0, gerror.Newf("unexpected bson type %s for float64", vr.Type().String())
    }
}

// Int32Decode 从bson值读取器中解码int32值
// vr bson值读取器
func Int32Decode(vr bson.ValueReader) (int32, error) {
    v, err := Int64Decode(vr) // 先解码为int64再转换
    return cast.ToInt32(v), err
}

// EnumDecode 从bson值读取器中解码枚举值
// vr bson值读取器
// T 枚举类型
func EnumDecode[T ~int32](vr bson.ValueReader) (T, error) {
    v, err := Int64Decode(vr) // 先解码为int64再转换
    return T(v), err
}

// Int64Decode 从bson值读取器中解码int64值
// vr bson值读取器
func Int64Decode(vr bson.ValueReader) (int64, error) {
    switch vr.Type() {
    case bson.TypeInt32:
        v, err := vr.ReadInt32()
        return cast.ToInt64(v), err
    case bson.TypeInt64:
        return vr.ReadInt64()
    case bson.TypeBoolean:
        v, err := vr.ReadBoolean()
        return cast.ToInt64(v), err
    case bson.TypeDouble:
        v, err := vr.ReadDouble()
        return cast.ToInt64(v), err
    case bson.TypeNull:
        return 0, vr.ReadNull()
    case bson.TypeUndefined:
        return 0, vr.ReadUndefined()
    default:
        return 0, gerror.Newf("unexpected bson type %s for int64", vr.Type().String())
    }
}

// UInt32Decode 从bson值读取器中解码uint32值
// vr bson值读取器
func UInt32Decode(vr bson.ValueReader) (uint32, error) {
    v, err := Int64Decode(vr) // 先解码为int64再转换
    return cast.ToUint32(v), err
}

// UInt64Decode 从bson值读取器中解码uint64值
// vr bson值读取器
func UInt64Decode(vr bson.ValueReader) (uint64, error) {
    v, err := Int64Decode(vr) // 先解码为int64再转换
    return cast.ToUint64(v), err
}

// BoolDecode 从bson值读取器中解码bool值
// vr bson值读取器
func BoolDecode(vr bson.ValueReader) (bool, error) {
    switch vr.Type() {
    case bson.TypeBoolean:
        return vr.ReadBoolean()
    case bson.TypeNull:
        return false, vr.ReadNull()
    case bson.TypeUndefined:
        return false, vr.ReadUndefined()
    default:
        return false, gerror.Newf("unexpected bson type %s for bool", vr.Type().String())
    }
}

// StringDecode 从bson值读取器中解码string值
// vr bson值读取器
func StringDecode(vr bson.ValueReader) (string, error) {
    switch vr.Type() {
    case bson.TypeString:
        return vr.ReadString()
    case bson.TypeNull:
        return "", vr.ReadNull()
    case bson.TypeUndefined:
        return "", vr.ReadUndefined()
    default:
        return "", gerror.Newf("unexpected bson type %s for string", vr.Type().String())
    }
}

// BinaryDecode 从bson值读取器中解码二进制数据
// vr bson值读取器
func BinaryDecode(vr bson.ValueReader) ([]byte, error) {
    switch vr.Type() {
    case bson.TypeBinary:
        v, _, err := vr.ReadBinary()
        return v, err
    case bson.TypeString:
        s, err := vr.ReadString()
        if err != nil {
            return nil, err
        }
        return base64.StdEncoding.DecodeString(s)
    case bson.TypeNull:
        return nil, vr.ReadNull()
    case bson.TypeUndefined:
        return nil, vr.ReadUndefined()
    default:
        return nil, gerror.Newf("unexpected bson type %s for binary", vr.Type().String())
    }
}
