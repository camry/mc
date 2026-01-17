package mc_test

import (
    "context"
    "testing"

    "github.com/camry/g/glog"
    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/bsonrw"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/camry/mc"
)

func defaultDB(names ...string) (*mc.MongoDatabase, error) {
    c, err := mc.NewMongoClient(
        context.Background(),
        options.Client().ApplyURI("mongodb://dev:Zqa5Q5pYgvmT@192.168.110.5:27017"),
    )
    if err != nil {
        return nil, err
    }
    name := "treasure-user"
    if len(names) > 0 {
        name = names[0]
    }
    return c.Database(name), nil
}

func TestNewMongoClient(t *testing.T) {
    db, err := defaultDB("treasure-user")
    assert.Nil(t, err)
    assert.Equal(t, "treasure-user", db.Name())

    adapter := mc.NewAdapter[*A, int64](db)
    a := NewA(adapter, 1)

    decode, err := mc.GetDecode(A{})
    assert.Nil(t, err)

    bsonRaw, err := db.Collection(a.Table()).FindOne(db.Context(), bson.M{"_id": a.Key()}).Raw()
    assert.Nil(t, err)

    vr := bsonrw.NewBSONDocumentReader(bsonRaw)

    err = decode.Decode(vr, a)
    assert.Nil(t, err)

    a.GetIds().ForEach(func(index int, item int32) {
        glog.Debugf("index: %d, item: %d", index, item)
    })

    a.GetB().GetC().GetD().GetTags().Append("tag1")
    a.GetB().GetC().GetD().GetTags().Append("tag2")
    a.GetB().GetC().GetD().GetTags().Append("tag3")
    a.GetB().GetC().GetD().GetTags().Append("tag4")
    a.GetB().GetC().GetD().GetTags().Append("tag5")
    a.GetB().GetC().GetD().GetTags().Remove("tag2")
    a.GetB().GetC().GetD().GetTags().Remove("tag4")

    // result, err := db.Collection(a.Table()).UpdateOne(c.Context(), bson.M{"_id": a.Key()}, bson.M{"$set": bson.M{"ids": a.GetIds().Marshal()}})
    // assert.Nil(t, err)
    // glog.Debugf("result: %+v", result)
}
