package mc

// DecodePair 解码对结构体，包含字段名称和对应的解码器
type DecodePair struct {
    name   string  // name 字段名称
    decode IDecode // decode 解码器
}

// NewDecodePair 创建新的解码对
// name 字段名称
// decode 解码器
func NewDecodePair(name string, decode IDecode) *DecodePair {
    return &DecodePair{
        name:   name,
        decode: decode,
    }
}
