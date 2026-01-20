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

type A struct {
    mc.MessageTop[*A, int64]

    _id       int64
    aUsername string
    ids       *mc.ListBase[int32]
    b         *B
}

var aDecode = mc.Register(&A{},
    mc.NewDecodePair("_id", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *A) (err error) {
        v._id, err = mc.Int64Decode(vr)
        return
    })),
    mc.NewDecodePair("aUsername", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *A) (err error) {
        v.aUsername, err = mc.StringDecode(vr)
        return
    })),
    mc.NewDecodePair("ids", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *A) (err error) {
        v.ids = mc.NewListBase[int32](v, "ids")
        err = v.ids.Decode(vr, v.ids)
        return
    })),
    mc.NewDecodePair("b", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *A) (err error) {
        v.b = newB(v, "b")
        err = bDecode.Decode(vr, v.b)
        return
    })),
)

func NewA(adapter mc.IAdapter[*A, int64], k int64) *A {
    x := &A{}
    x._id = k
    x.Bind(adapter, x, k)
    return x
}

func (a *A) Table() string {
    return "users"
}

func (a *A) Marshal() bson.M {
    m := make(bson.M)
    m["_id"] = a.Key()
    m["aUsername"] = a.GetAUsername()
    m["ids"] = a.GetIds().Marshal()
    m["b"] = a.GetB().Marshal()
    return m
}

func (a *A) GetAUsername() string {
    return a.aUsername
}

func (a *A) SetAUsername(username string) {
    a.aUsername = username
    a.Set(a, "aUsername", username)
}

func (a *A) GetIds() *mc.ListBase[int32] {
    if a.ids == nil {
        a.ids = mc.NewListBase[int32](a, "ids")
    }
    return a.ids
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

var bDecode = mc.Register(&B{},
    mc.NewDecodePair("bUsername", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *B) (err error) {
        v.bUsername, err = mc.StringDecode(vr)
        return
    })),
    mc.NewDecodePair("c", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *B) (err error) {
        v.c = newC(v, "c")
        err = cDecode.Decode(vr, v.c)
        return
    })),
)

func newB(parent mc.IMessage, key any) *B {
    x := &B{}
    x.Bind(parent, key)
    return x
}

func (b *B) Marshal() bson.M {
    m := make(bson.M)
    m["bUsername"] = b.GetBUsername()
    m["c"] = b.GetC().Marshal()
    return m
}

func (b *B) GetBUsername() string {
    return b.bUsername
}

func (b *B) SetBUsername(username string) {
    b.bUsername = username
    b.Set(b, "bUsername", username)
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
    dList     *mc.ListMessage[*D]
    tagMap    *mc.MapBase[string, int32]
    dMap      *mc.MapMessage[string, *D]
}

var cDecode = mc.Register(&C{},
    mc.NewDecodePair("cUsername", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *C) (err error) {
        v.cUsername, err = mc.StringDecode(vr)
        return
    })),
    mc.NewDecodePair("d", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *C) (err error) {
        v.d = newD(v, "d")
        err = dDecode.Decode(vr, v.d)
        return
    })),
    mc.NewDecodePair("dList", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *C) (err error) {
        v.dList = mc.NewListMessage[*D](v, "dList", newD)
        err = v.dList.Decode(vr, v.dList)
        return
    })),
    mc.NewDecodePair("tagMap", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *C) (err error) {
        v.tagMap = mc.NewMapBase[string, int32](v, "tagMap")
        err = v.tagMap.Decode(vr, v.tagMap)
        return
    })),
    mc.NewDecodePair("dMap", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *C) (err error) {
        v.dMap = mc.NewMapMessage[string, *D](v, "dMap", newD)
        err = v.dMap.Decode(vr, v.dMap)
        return
    })),
)

func (c *C) Marshal() bson.M {
    m := make(bson.M)
    m["cUsername"] = c.GetCUsername()
    m["d"] = c.GetD().Marshal()
    m["dList"] = c.GetDList().Marshal()
    m["tagMap"] = c.GetTagMap().Marshal()
    m["dMap"] = c.GetDMap().Marshal()
    return m
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
    c.Set(c, "cUsername", username)
}

func (c *C) GetD() *D {
    if c.d == nil {
        c.d = newD(c, "d")
    }
    return c.d
}

func (c *C) GetDList() *mc.ListMessage[*D] {
    if c.dList == nil {
        c.dList = mc.NewListMessage[*D](c, "dList", newD)
    }
    return c.dList
}

func (c *C) GetTagMap() *mc.MapBase[string, int32] {
    if c.tagMap == nil {
        c.tagMap = mc.NewMapBase[string, int32](c, "tagMap")
    }
    return c.tagMap
}

func (c *C) GetDMap() *mc.MapMessage[string, *D] {
    if c.dMap == nil {
        c.dMap = mc.NewMapMessage[string, *D](c, "dMap", newD)
    }
    return c.dMap
}

type D struct {
    mc.MessageSub

    dUsername string
    tags      *mc.ListBase[string]
}

var dDecode = mc.Register(&D{},
    mc.NewDecodePair("dUsername", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *D) (err error) {
        v.dUsername, err = mc.StringDecode(vr)
        return
    })),
    mc.NewDecodePair("tags", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *D) (err error) {
        v.tags = mc.NewListBase[string](v, "tags")
        err = v.tags.Decode(vr, v.tags)
        return
    })),
)

func (d *D) Marshal() bson.M {
    m := make(bson.M)
    m["dUsername"] = d.GetDUsername()
    m["tags"] = d.GetTags().Marshal()
    return m
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
    d.Set(d, "dUsername", username)
}

func (d *D) GetTags() *mc.ListBase[string] {
    if d.tags == nil {
        d.tags = mc.NewListBase[string](d, "tags")
    }
    return d.tags
}

func TestAdapter_Set(t *testing.T) {
    db, err := defaultDB("treasure-user")
    assert.Nil(t, err)
    assert.Equal(t, "treasure-user", db.Name())

    adapter := mc.NewAdapter[*A, int64](db)
    a := NewA(adapter, 2)

    decode, err := mc.GetDecode(&A{})
    assert.Nil(t, err)
    bsonRaw, err := db.Collection(a.Table()).FindOne(db.Context(), bson.M{"_id": a.Key()}).Raw()
    if errors.Is(err, mongo.ErrNoDocuments) {
        db.Collection(a.Table()).InsertOne(db.Context(), a.Marshal())
    } else {
        assert.Nil(t, err)
    }
    err = decode.Decode(bsonrw.NewBSONDocumentReader(bsonRaw), a)
    assert.Nil(t, err)

    glog.Infof("Marshal %v", a.Marshal())
}
