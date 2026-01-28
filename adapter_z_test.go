package mc_test

import (
    "math"
    "testing"
    "time"

    "github.com/camry/g/glog"
    "github.com/stretchr/testify/assert"

    "github.com/camry/mc"
    "github.com/camry/mc/pb"
)

func TestNewAdapter(t *testing.T) {
    db, err := defaultDB("treasure-user")
    assert.Nil(t, err)
    assert.Equal(t, "treasure-user", db.Name())

    adapter := mc.NewAdapter[*pb.A, int64](db, mc.WithAdapterEnableSave[*pb.A, int64](true))
    a := pb.NewA(adapter, 2)

    err = db.Find(a)
    assert.Nil(t, err)

    a.SetAUsername("camry")
    a.SetANicknameAlias("Camry.Chen")
    a.GetB().GetC().GetD().SetId(math.MaxInt32)
    a.GetB().GetC().GetD().SetIid(math.MaxUint32)
    a.GetB().GetC().GetD().SetUid(math.MaxInt64)
    a.GetB().GetC().GetD().SetDUsername("camry")
    a.GetB().GetC().GetD().SetIsBool(true)
    a.GetB().GetC().GetD().SetFltV32(1234.1234)
    a.GetB().GetC().GetD().SetFltV64(1234.123456789)
    a.GetB().GetC().GetD().SetBytes([]byte("hello"))
    a.GetB().GetC().GetD().GetTags().Append("tag3")
    a.GetB().GetC().GetD().GetTags().Remove("tag3")

    time.Sleep(time.Second)
    glog.Infof("Marshal %v", a.Marshal())
}
