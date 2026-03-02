package mc_test

import (
    "context"

    "go.mongodb.org/mongo-driver/v2/mongo/options"

    "gitea.meigushi.go/meigushi/mc"
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
