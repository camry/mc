package mc_test

import (
    "errors"
    "testing"

    "github.com/camry/g/glog"
    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
    "go.mongodb.org/mongo-driver/mongo"

    "github.com/camry/mc"
)

func TestNewAdapter(t *testing.T) {
    db, err := defaultDB("treasure-user")
    assert.Nil(t, err)
    assert.Equal(t, "treasure-user", db.Name())

    adapter := mc.NewAdapter[*A, int64](db)
    a := NewA(adapter, 2)

    decode, err := mc.GetDecode(&A{})
    assert.Nil(t, err)
    bsonRaw, err := db.Collection(a.Table()).FindOne(db.Context(), bson.M{"_id": a.Key()}).Raw()
    if errors.Is(err, mongo.ErrNoDocuments) {
        _, err1 := db.Collection(a.Table()).InsertOne(db.Context(), a.Marshal())
        assert.Nil(t, err1)
    } else {
        assert.Nil(t, err)
    }
    err = decode.Decode(bsonrw.NewBSONDocumentReader(bsonRaw), a)
    assert.Nil(t, err)

    glog.Infof("Marshal %v", a.Marshal())
}
