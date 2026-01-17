package mc

import (
    "errors"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

type DecodeMessage struct {
    fieldMap map[string]IDecode
}

func NewDecodeMessage(decodePairs ...*DecodePair) *DecodeMessage {
    d := &DecodeMessage{
        fieldMap: make(map[string]IDecode),
    }
    for _, decodePair := range decodePairs {
        d.fieldMap[decodePair.name] = decodePair.decode
    }
    return d
}

func (x *DecodeMessage) Decode(vr bsonrw.ValueReader, v any) error {
    if vr.Type() == bson.TypeNull {
        return vr.ReadNull()
    }
    if vr.Type() == bson.TypeUndefined {
        return vr.ReadUndefined()
    }

    dr, err := vr.ReadDocument()
    if err != nil {
        return err
    }

    for {
        key, evr, err1 := dr.ReadElement()
        if errors.Is(err1, bsonrw.ErrEOD) {
            break
        }
        if err1 != nil {
            return err1
        }
        handler, ok := x.fieldMap[key]
        if ok {
            if err1 = handler.Decode(evr, v); err1 != nil {
                return err1
            }
        } else {
            if err1 = evr.Skip(); err1 != nil {
                return err1
            }
        }
    }

    return nil
}
