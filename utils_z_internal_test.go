package mc

import (
    "reflect"
    "testing"
)

func TestToMapKey(t *testing.T) {
    if k, err := toMapKey[int64]("42", reflect.Int64); err != nil || k != 42 {
        t.Fatalf("int64: got %v, %v", k, err)
    }
    if k, err := toMapKey[uint32]("4294967295", reflect.Uint32); err != nil || k != 4294967295 {
        t.Fatalf("uint32 max: got %v, %v", k, err)
    }
    if k, err := toMapKey[string]("abc", reflect.String); err != nil || k != "abc" {
        t.Fatalf("string: got %v, %v", k, err)
    }
    if _, err := toMapKey[int64]("abc", reflect.Int64); err == nil {
        t.Fatal("expected error for non-numeric int64 key")
    }
    if _, err := toMapKey[uint32]("-1", reflect.Uint32); err == nil {
        t.Fatal("expected error for negative uint32 key")
    }
}
