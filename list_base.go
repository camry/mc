package mc

import (
    "errors"
    "reflect"
    "slices"

    "github.com/camry/g/gerrors/gerror"
    "github.com/samber/lo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
    "go.mongodb.org/mongo-driver/bson/bsontype"
)

type ListBase[T ListBaseType] struct {
    List[T]
}

func NewListBase[T ListBaseType](parent IMessage, key string) *ListBase[T] {
    return &ListBase[T]{
        List: List[T]{
            parent: parent,
            key:    key,
        },
    }
}

func (x *ListBase[T]) MessageType() MessageType {
    return MessageTypeListBase
}

func (x *ListBase[T]) Key() any {
    return x.key
}

func (x *ListBase[T]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Set(x.parent, key, val, nestedMessages...)
}

func (x *ListBase[T]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)
    x.parent.Unset(x.parent, key, nestedMessages...)
}

func (x *ListBase[T]) Remove(item T) {
    index := slices.Index(x.v, item)
    if index == -1 {
        return
    }
    x.List.Remove(index)
}

func (x *ListBase[T]) Decode(vr bsonrw.ValueReader, _ any) error {
    if vr.Type() == bson.TypeArray {
        arr, err := vr.ReadArray()
        if err != nil {
            return err
        }
        for {
            avr, err1 := arr.ReadValue()
            if errors.Is(err1, bsonrw.ErrEOA) {
                break
            }
            if err1 != nil {
                return err1
            }
            if !lo.Contains([]bsontype.Type{
                bson.TypeInt32, bson.TypeInt64, bson.TypeDouble, bson.TypeString,
            }, avr.Type()) {
                continue
            }
            var (
                val  any
                err2 error
            )
            kind := reflect.TypeOf(x.v).Elem().Kind()
            switch kind {
            case reflect.Int32:
                val, err2 = avr.ReadInt32()
            case reflect.Int64:
                val, err2 = avr.ReadInt64()
            case reflect.Float32:
                vv, err3 := avr.ReadDouble()
                if err3 != nil {
                    return err3
                }
                val = float32(vv)
            case reflect.Float64:
                val, err2 = avr.ReadDouble()
            case reflect.String:
                val, err2 = avr.ReadString()
            default:
                err2 = gerror.Newf("unknow ListBaseType %s", kind)
            }
            if err2 != nil {
                return err2
            }
            x.v = append(x.v, val.(T))
        }
    }
    return nil
}
