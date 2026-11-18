package mc_test

import (
    "sync"
    "sync/atomic"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/camry/mc"
)

// testParent 模拟父级消息对象
type testParent struct{}

func (testParent) MessageType() mc.MessageType             { return mc.MessageTypeTop }
func (testParent) Key() any                                { return nil }
func (testParent) Set(_ any, _ any, _ ...mc.IMessage)       {}
func (testParent) Unset(_ any, _ ...mc.IMessage)            {}

// testMessage 模拟 MapMessage 的值对象
type testMessage struct {
    id int64
}

func (x *testMessage) MessageType() mc.MessageType           { return mc.MessageTypeSub }
func (x *testMessage) Key() any                              { return x.id }
func (x *testMessage) Set(_ any, _ any, _ ...mc.IMessage)     {}
func (x *testMessage) Unset(_ any, _ ...mc.IMessage)          {}

func TestMapBaseConcurrentReadWrite(t *testing.T) {
    m := mc.NewMapBase[int64, int64](testParent{}, "m")

    const goroutines = 16
    const rounds = 200

    var wg sync.WaitGroup
    for i := range goroutines {
        wg.Go(func() {
            for j := range rounds {
                key := int64(i*rounds + j)
                m.Put(key, key)
                _, _ = m.Get(key)
                _ = m.Contains(key)
                _ = m.Size()
                _ = m.IsEmpty()
                _ = m.Keys()
                _ = m.Values()
                _ = m.GetAll()
                _ = m.Marshal()
                m.Range(func(_ int64, _ int64) bool { return true })
                if j%2 == 0 {
                    m.Delete(key)
                }
            }
        })
    }
    wg.Wait()

    // 偶数轮被删除，奇数轮保留
    assert.Equal(t, int64(goroutines*(rounds/2)), int64(m.Size()))
}

func TestMapBaseClearConcurrent(t *testing.T) {
    m := mc.NewMapBase[int64, int64](testParent{}, "m")

    var wg sync.WaitGroup
    for i := range 8 {
        wg.Go(func() {
            for j := range 500 {
                m.Put(int64(i*500+j), 1)
                if j%100 == 0 {
                    m.Clear()
                }
            }
        })
    }
    wg.Wait()

    // 不 panic、不触发 race 即可；最终大小取决于调度，只做范围断言
    assert.GreaterOrEqual(t, m.Size(), 0)
}

func TestMapBaseIncrConcurrent(t *testing.T) {
    m := mc.NewMapBase[int64, int64](testParent{}, "m")

    const goroutines = 32
    const rounds = 1000

    var wg sync.WaitGroup
    for range goroutines {
        wg.Go(func() {
            for range rounds {
                mc.MapBaseIncr(m, int64(1), int64(1))
            }
        })
    }
    wg.Wait()

    v, ok := m.Get(1)
    assert.True(t, ok)
    assert.Equal(t, int64(goroutines*rounds), v)
}

func TestMapMessageGetOrPutConcurrent(t *testing.T) {
    var created atomic.Int64
    creator := func(_ mc.IMessage, key any) *testMessage {
        created.Add(1)
        return &testMessage{id: key.(int64)}
    }
    m := mc.NewMapMessage[int64, *testMessage](testParent{}, "m", creator)

    const goroutines = 32
    results := make([]*testMessage, goroutines)

    var wg sync.WaitGroup
    for i := range goroutines {
        wg.Go(func() {
            results[i] = mc.MapMessageGetOrPut(m, 7)
        })
    }
    wg.Wait()

    assert.Equal(t, int64(1), created.Load()) // 同 key 只创建一个对象
    for _, r := range results {
        assert.Same(t, results[0], r)
    }
}

func TestMapRangeMayWrite(t *testing.T) {
    m := mc.NewMapBase[int64, int64](testParent{}, "m")
    m.Put(1, 1)
    m.Put(2, 2)

    // Range 回调中写 map 不应死锁
    m.Range(func(k int64, _ int64) bool {
        m.Put(k+100, k)
        m.Delete(k)
        return true
    })

    assert.Equal(t, 2, m.Size())
    assert.True(t, m.Contains(101))
    assert.True(t, m.Contains(102))
}
