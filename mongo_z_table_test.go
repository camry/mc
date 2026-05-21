package mc_test

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "gitea.meigushi.go/meigushi/mc"
    "gitea.meigushi.go/meigushi/mc/pb"
)

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
