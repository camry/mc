package mc_test

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "gitea.meigushi.go/meigushi/mc"
    "gitea.meigushi.go/meigushi/mc/pb"
)

func TestMongoTable_FindOne(t *testing.T) {
    db, err := defaultDB("treasure-user")
    assert.Nil(t, err)
    assert.Equal(t, "treasure-user", db.Name())

    a1, err := mc.NewMongoTable[*pb.A, int64](db).FindOne(1)
    assert.Nil(t, err)
    a2, err := mc.NewMongoTable[*pb.A, int64](db).FindOne(2)
    assert.Nil(t, err)

    assert.Equal(t, a1.GetXId(), int64(1))
    assert.Equal(t, a2.GetXId(), int64(2))
}

func TestMongoTable_FindMany(t *testing.T) {
    db, err := defaultDB("treasure-user")
    assert.Nil(t, err)
    assert.Equal(t, "treasure-user", db.Name())

    ids := []int64{1, 2}
    lst, err := mc.NewMongoTable[*pb.A, int64](db).FindMany(ids)
    assert.Nil(t, err)

    for _, a := range lst {
        t.Log(a.Marshal())
    }
}
