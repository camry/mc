package mc_test

import (
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

    adapter := mc.NewAdapter[*pb.A, int64](db)
    a := pb.NewA(adapter, 2)

    err = db.Find(a)
    assert.Nil(t, err)

    a.SetAUsername("camry")
    a.SetANicknameAlias("Camry.Chen")

    time.Sleep(time.Second)
    glog.Infof("Marshal %v", a.Marshal())
}
