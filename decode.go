package mc

import (
    "github.com/camry/g/gerrors/gerror"
)

var decodeMap = make(map[string]IDecode)

func Register(message IMessage, decodePairs ...*DecodePair) *DecodeMessage {
    decodeMessage := NewDecodeMessage(decodePairs...)
    decodeMap[GetStructName(message)] = decodeMessage
    return decodeMessage
}

func GetDecode(v any) (IDecode, error) {
    if decode, ok := decodeMap[GetStructName(v)]; ok {
        return decode, nil
    }
    return nil, gerror.Newf("no decode for %s", GetStructName(v))
}
