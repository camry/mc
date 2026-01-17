package mc

import (
    "github.com/camry/g/gerrors/gerror"
    "github.com/spf13/cast"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

func Float32Decode(vr bsonrw.ValueReader) (float32, error) {
    v, err := Float64Decode(vr)
    return cast.ToFloat32(v), err
}

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
        return vr.ReadDouble()
    case bson.TypeNull:
        return 0, vr.ReadNull()
    case bson.TypeUndefined:
        return 0, vr.ReadUndefined()
    default:
        return 0, gerror.Newf("unexpected bson type %s for float64", vr.Type().String())
    }
}

func Int32Decode(vr bsonrw.ValueReader) (int32, error) {
    v, err := Int64Decode(vr)
    return cast.ToInt32(v), err
}

func Int64Decode(vr bsonrw.ValueReader) (int64, error) {
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

func UInt32Decode(vr bsonrw.ValueReader) (uint32, error) {
    v, err := Int64Decode(vr)
    return cast.ToUint32(v), err
}

func UInt64Decode(vr bsonrw.ValueReader) (uint64, error) {
    v, err := Int64Decode(vr)
    return cast.ToUint64(v), err
}

func BoolDecode(vr bsonrw.ValueReader) (bool, error) {
    switch vr.Type() {
    case bson.TypeBoolean:
        return vr.ReadBoolean()
    default:
        return false, gerror.Newf("unexpected bson type %s for bool", vr.Type().String())
    }
}

func StringDecode(vr bsonrw.ValueReader) (string, error) {
    switch vr.Type() {
    case bson.TypeString:
        return vr.ReadString()
    default:
        return "", gerror.Newf("unexpected bson type %s for string", vr.Type().String())
    }
}

func BinaryDecode(vr bsonrw.ValueReader) ([]byte, error) {
    switch vr.Type() {
    case bson.TypeBinary:
        v, _, err := vr.ReadBinary()
        return v, err
    default:
        return nil, gerror.Newf("unexpected bson type %s for binary", vr.Type().String())
    }
}
