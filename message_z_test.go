package mc_test

import (
    "testing"

    "github.com/camry/mc"
)

type A struct {
    mc.MessageTop[int64]

    aUsername string
    b         *B
}

func newA(adapter mc.IAdapter[int64], k int64) *A {
    x := &A{}
    x.Bind(adapter, k)
    return x
}

func (a *A) GetAUsername() string {
    return a.aUsername
}

func (a *A) SetAUsername(username string) {
    a.aUsername = username
    a.Collect(a, "aUsername", username)
}

func (a *A) GetB() *B {
    if a.b == nil {
        a.b = newB(a, "b")
    }
    return a.b
}

type B struct {
    mc.MessageSub

    bUsername string
    c         *C
}

func newB(parent mc.IMessage, key any) *B {
    x := &B{}
    x.Bind(parent, key)
    return x
}

func (b *B) GetBUsername() string {
    return b.bUsername
}

func (b *B) SetBUsername(username string) {
    b.bUsername = username
    b.Collect(b, "bUsername", username)
}

func (b *B) GetC() *C {
    if b.c == nil {
        b.c = newC(b, "c")
    }
    return b.c
}

type C struct {
    mc.MessageSub

    cUsername string
    d         *D
}

func newC(parent mc.IMessage, key any) *C {
    x := &C{}
    x.Bind(parent, key)
    return x
}

func (c *C) GetCUsername() string {
    return c.cUsername
}

func (c *C) SetCUsername(username string) {
    c.cUsername = username
    c.Collect(c, "cUsername", username)
}

func (c *C) GetD() *D {
    if c.d == nil {
        c.d = newD(c, "d")
    }
    return c.d
}

type D struct {
    mc.MessageSub

    dUsername string
}

func newD(parent mc.IMessage, key any) *D {
    x := &D{}
    x.Bind(parent, key)
    return x
}

func (d *D) GetDUsername() string {
    return d.dUsername
}

func (d *D) SetDUsername(username string) {
    d.dUsername = username
    d.Collect(d, "dUsername", username)
}

func TestAdapter_Collect(t *testing.T) {
    adapter := mc.NewAdapter[int64](1)
    a := newA(adapter, 1)
    a.GetB().GetC().GetD().SetDUsername("camry")
}
