package mc

import (
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

type ListMessage[T IMessage] struct {
    List[T]
}

func NewListMessage[T IMessage](parent IMessage, key string) *ListMessage[T] {
    return &ListMessage[T]{
        List: List[T]{
            parent: parent,
            key:    key,
        },
    }
}

func (x *ListMessage[T]) Decode(vr bsonrw.ValueReader, _ any) error {
    if vr.Type() == bson.TypeArray {
        // arr, err := vr.ReadArray()
        // if err != nil {
        //     return err
        // }
        // for {
        //     avr, err1 := arr.ReadValue()
        //     if errors.Is(err1, bsonrw.ErrEOA) {
        //         break
        //     }
        //     if err1 != nil {
        //         return err1
        //     }
        //     if avr.Type() != bson.TypeEmbeddedDocument {
        //         continue
        //     }
        //     var (
        //         val T
        //     )
        //     decodeMessage, err1 := GetDecode(val)
        //     if err1 != nil {
        //         return err1
        //     }
        //     ptrVal := new(T)
        //     val = *ptrVal
        //     // TODO 通过 reflect 实例化 val 对象
        //     err1 = decodeMessage.Decode(avr, val)
        //     if err1 != nil {
        //         return err1
        //     }
        //     x.v = append(x.v, val)
        // }
    }
    return nil
}
