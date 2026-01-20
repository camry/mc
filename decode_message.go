package mc

import (
    "errors"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
)

// DecodeMessage 消息解码器结构体
type DecodeMessage struct {
    // fieldMap 字段名称与解码器的映射
    fieldMap map[string]IDecode
}

// NewDecodeMessage 创建新的消息解码器
// decodePairs 解码对数组
func NewDecodeMessage(decodePairs ...*DecodePair) *DecodeMessage {
    d := &DecodeMessage{
        fieldMap: make(map[string]IDecode), // 初始化字段映射
    }
    for _, decodePair := range decodePairs {
        d.fieldMap[decodePair.name] = decodePair.decode // 将解码对加入映射
    }
    return d
}

// Decode 执行消息解码操作
// vr bson值读取器
// v 解码目标对象
func (x *DecodeMessage) Decode(vr bsonrw.ValueReader, v any) error {
    // 检查值类型是否为空或未定义
    if vr.Type() == bson.TypeNull {
        return vr.ReadNull()
    }
    if vr.Type() == bson.TypeUndefined {
        return vr.ReadUndefined()
    }

    // 开始读取文档
    dr, err := vr.ReadDocument()
    if err != nil {
        return err
    }

    // 遍历文档中的每个元素
    for {
        key, evr, err1 := dr.ReadElement()
        if errors.Is(err1, bsonrw.ErrEOD) { // 到达文档末尾
            break
        }
        if err1 != nil {
            return err1
        }
        // 查找对应的解码处理器
        handler, ok := x.fieldMap[key]
        if ok {
            if err1 = handler.Decode(evr, v); err1 != nil { // 执行解码
                return err1
            }
        } else {
            if err1 = evr.Skip(); err1 != nil { // 跳过未处理的元素
                return err1
            }
        }
    }

    return nil
}
