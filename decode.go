package mc

import (
    "github.com/camry/g/gerrors/gerror"
)

// decodeMap 存储消息类型与解码器的映射关系
var decodeMap = make(map[string]IDecode)

// Register 注册消息类型及其对应的解码对
// message 要注册的消息对象
// decodePairs 解码对数组
func Register(message IMessage, decodePairs ...*DecodePair) *DecodeMessage {
    decodeMessage := NewDecodeMessage(decodePairs...) // 创建解码消息对象
    decodeMap[GetStructName(message)] = decodeMessage // 将解码消息存入映射
    return decodeMessage
}

// GetDecode 根据消息类型获取对应的解码器
// v 消息对象
func GetDecode(v any) (IDecode, error) {
    if decode, ok := decodeMap[GetStructName(v)]; ok {
        return decode, nil
    }
    return nil, gerror.Newf("no decode for %s", GetStructName(v))
}
