package mc

type DecodePair struct {
    name   string
    decode IDecode
}

func NewDecodePair(name string, decode IDecode) *DecodePair {
    return &DecodePair{
        name:   name,
        decode: decode,
    }
}
