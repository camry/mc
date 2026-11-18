package mc_test

import (
    "bytes"
    "math"
    "testing"

    "go.mongodb.org/mongo-driver/v2/bson"

    "github.com/camry/mc"
)

func newValueReader(t *testing.T, value any) bson.ValueReader {
    t.Helper()
    raw, err := bson.Marshal(bson.M{"v": value})
    if err != nil {
        t.Fatalf("marshal: %v", err)
    }
    dr, err := bson.NewDocumentReader(bytes.NewReader(raw)).ReadDocument()
    if err != nil {
        t.Fatalf("read document: %v", err)
    }
    _, vr, err := dr.ReadElement()
    if err != nil {
        t.Fatalf("read element: %v", err)
    }
    return vr
}

func TestInt32DecodeRange(t *testing.T) {
    if v, err := mc.Int32Decode(newValueReader(t, int32(math.MaxInt32))); err != nil || v != math.MaxInt32 {
        t.Fatalf("max int32: got %v, %v", v, err)
    }
    if _, err := mc.Int32Decode(newValueReader(t, int64(math.MaxInt32)+1)); err == nil {
        t.Fatal("expected overflow error for int64 > MaxInt32")
    }
    if _, err := mc.Int32Decode(newValueReader(t, int64(math.MinInt32)-1)); err == nil {
        t.Fatal("expected overflow error for int64 < MinInt32")
    }
}

func TestUIntDecodeRange(t *testing.T) {
    if v, err := mc.UInt32Decode(newValueReader(t, int64(math.MaxUint32))); err != nil || v != math.MaxUint32 {
        t.Fatalf("max uint32: got %v, %v", v, err)
    }
    if _, err := mc.UInt32Decode(newValueReader(t, int64(-1))); err == nil {
        t.Fatal("expected error for negative uint32")
    }
    if _, err := mc.UInt64Decode(newValueReader(t, int64(-1))); err == nil {
        t.Fatal("expected error for negative uint64")
    }
    if v, err := mc.UInt64Decode(newValueReader(t, int64(42))); err != nil || v != 42 {
        t.Fatalf("uint64: got %v, %v", v, err)
    }
}

func TestInt64DecodeDoubleRange(t *testing.T) {
    if v, err := mc.Int64Decode(newValueReader(t, 1e19)); err == nil {
        t.Fatalf("expected overflow error for 1e19, got %v", v)
    }
    if _, err := mc.Int64Decode(newValueReader(t, math.NaN())); err == nil {
        t.Fatal("expected error for NaN")
    }
    if v, err := mc.Int64Decode(newValueReader(t, 123.0)); err != nil || v != 123 {
        t.Fatalf("integral double: got %v, %v", v, err)
    }
}

func TestFloat32DecodeRange(t *testing.T) {
    if _, err := mc.Float32Decode(newValueReader(t, 1e40)); err == nil {
        t.Fatal("expected overflow error for 1e40 -> float32")
    }
    if v, err := mc.Float32Decode(newValueReader(t, 1.5)); err != nil || v != 1.5 {
        t.Fatalf("float32: got %v, %v", v, err)
    }
}
