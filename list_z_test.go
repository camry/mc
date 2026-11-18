package mc_test

import (
    "sync"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/camry/mc"
)

func TestListBaseConcurrentAppend(t *testing.T) {
    l := mc.NewListBase[int64](testParent{}, "l")

    const goroutines = 16
    const rounds = 200

    var wg sync.WaitGroup
    for i := range goroutines {
        wg.Go(func() {
            for j := range rounds {
                l.Append(int64(i*rounds + j))
            }
        })
    }
    wg.Wait()

    assert.Equal(t, goroutines*rounds, l.Len())
}

func TestListBaseConcurrentMixed(t *testing.T) {
    l := mc.NewListBase[int64](testParent{}, "l")

    var wg sync.WaitGroup
    for i := range 8 {
        wg.Go(func() {
            for j := range 300 {
                v := int64(i*300 + j)
                l.Append(v)
                _ = l.Len()
                _ = l.IsEmpty()
                _ = l.Contains(v)
                _ = l.Values()
                _ = l.Marshal()
                _ = l.Filter(func(x int64) bool { return x%2 == 0 })
                _, _ = l.Find(func(x int64) bool { return x == v })
                l.Range(func(_ int, _ int64) bool { return true })
                if j%3 == 0 {
                    l.Remove(v)
                }
            }
        })
    }
    wg.Wait()

    // 不 panic、不丢结构即可；保留数量为未被删除的元素
    assert.Equal(t, 8*300*2/3, l.Len())
}

func TestListBaseRemoveValueConcurrent(t *testing.T) {
    l := mc.NewListBase[int64](testParent{}, "l")
    const n = 2000
    for i := range n {
        l.Append(int64(i))
    }

    var wg sync.WaitGroup
    for i := range n {
        wg.Go(func() {
            l.Remove(int64(i))
        })
    }
    wg.Wait()

    assert.Equal(t, 0, l.Len())
}

func TestListMessageAppendNewMessageConcurrent(t *testing.T) {
    creator := func(_ mc.IMessage, key any) *testMessage {
        return &testMessage{id: int64(key.(int))}
    }
    l := mc.NewListMessage[*testMessage](testParent{}, "l", creator)

    const goroutines = 32
    const rounds = 50

    var wg sync.WaitGroup
    for range goroutines {
        wg.Go(func() {
            for range rounds {
                l.AppendNewMessage()
            }
        })
    }
    wg.Wait()

    values := l.Values()
    assert.Len(t, values, goroutines*rounds)

    seen := make(map[int64]bool, len(values))
    for _, v := range values {
        assert.False(t, seen[v.id], "duplicate key %d", v.id)
        seen[v.id] = true
    }
    for i := range goroutines * rounds {
        assert.True(t, seen[int64(i)], "missing key %d", i)
    }
}

func TestListRangeMayWrite(t *testing.T) {
    l := mc.NewListBase[int64](testParent{}, "l")
    l.Append(1)
    l.Append(2)

    // Range 回调中写列表不应死锁
    l.Range(func(_ int, v int64) bool {
        l.Append(v + 100)
        return true
    })

    assert.Equal(t, 4, l.Len())
    assert.True(t, l.Contains(101))
    assert.True(t, l.Contains(102))
}

func TestListValuesReturnsCopy(t *testing.T) {
    l := mc.NewListBase[int64](testParent{}, "l")
    l.Append(1)
    l.Append(2)

    values := l.Values()
    values[0] = 100

    assert.Equal(t, int64(1), l.Values()[0])
}
