package mc

import (
    "github.com/camry/g/gerrors/gerror"
    "github.com/spf13/cast"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

// Float32Decode 从bson值读取器中解码float32值
// vr bson值读取器
func Float32Decode(vr bsonrw.ValueReader) (float32, error) {
    v, err := Float64Decode(vr) // 先解码为float64再转换
    return cast.ToFloat32(v), err
}

// Float64Decode 从bson值读取器中解码float64值
// vr bson值读取器
func Float64Decode(vr bsonrw.ValueReader) (float64, error) {
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
        return vr.ReadDouble() // 直接读取double
    case bson.TypeNull:
        return 0, vr.ReadNull() // 读取null值
    case bson.TypeUndefined:
        return 0, vr.ReadUndefined() // 读取undefined值
    default:
        return 0, gerror.Newf("unexpected bson type %s for float64", vr.Type().String()) // 不支持的类型
    }
}

// Int32Decode 从bson值读取器中解码int32值
// vr bson值读取器
func Int32Decode(vr bsonrw.ValueReader) (int32, error) {
    v, err := Int64Decode(vr) // 先解码为int64再转换
    return cast.ToInt32(v), err
}

// Int64Decode 从bson值读取器中解码int64值
// vr bson值读取器
func Int64Decode(vr bsonrw.ValueReader) (int64, error) {
    switch vr.Type() {
    case bson.TypeInt32:
        v, err := vr.ReadInt32()
        return cast.ToInt64(v), err
    case bson.TypeInt64:
        return vr.ReadInt64() // 直接读取int64
    case bson.TypeBoolean:
        v, err := vr.ReadBoolean()
        return cast.ToInt64(v), err
    case bson.TypeDouble:
        v, err := vr.ReadDouble()
        return cast.ToInt64(v), err
    case bson.TypeNull:
        return 0, vr.ReadNull() // 读取null值
    case bson.TypeUndefined:
        return 0, vr.ReadUndefined() // 读取undefined值
    default:
        return 0, gerror.Newf("unexpected bson type %s for int64", vr.Type().String()) // 不支持的类型
    }
}

// UInt32Decode 从bson值读取器中解码uint32值
// vr bson值读取器
func UInt32Decode(vr bsonrw.ValueReader) (uint32, error) {
    v, err := Int64Decode(vr) // 先解码为int64再转换
    return cast.ToUint32(v), err
}

// UInt64Decode 从bson值读取器中解码uint64值
// vr bson值读取器
func UInt64Decode(vr bsonrw.ValueReader) (uint64, error) {
    v, err := Int64Decode(vr) // 先解码为int64再转换
    return cast.ToUint64(v), err
}

// BoolDecode 从bson值读取器中解码bool值
// vr bson值读取器
func BoolDecode(vr bsonrw.ValueReader) (bool, error) {
    switch vr.Type() {
    case bson.TypeBoolean:
        return vr.ReadBoolean() // 直接读取布尔值
    default:
        return false, gerror.Newf("unexpected bson type %s for bool", vr.Type().String()) // 不支持的类型
    }
}

// StringDecode 从bson值读取器中解码string值
// vr bson值读取器
func StringDecode(vr bsonrw.ValueReader) (string, error) {
    switch vr.Type() {
    case bson.TypeString:
        return vr.ReadString() // 直接读取字符串
    default:
        return "", gerror.Newf("unexpected bson type %s for string", vr.Type().String()) // 不支持的类型
    }
}

// BinaryDecode 从bson值读取器中解码二进制数据
// vr bson值读取器
func BinaryDecode(vr bsonrw.ValueReader) ([]byte, error) {
    switch vr.Type() {
    case bson.TypeBinary:
        v, _, err := vr.ReadBinary() // 直接读取二进制数据
        return v, err
    default:
        return nil, gerror.Newf("unexpected bson type %s for binary", vr.Type().String()) // 不支持的类型
    }
}
