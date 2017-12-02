package main

import (
    "fmt"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

type Key struct {
    Name string
    Value string
}

type KeyStore struct {
    session *mgo.Session
}

func (ks *KeyStore) put(key *Key) {
    var coll = ks.session.Copy().DB(config.mongoDb).C(config.mongoColl)
    err := coll.Insert(key)
    if err != nil {
        fmt.Println("Error put key: ", err.Error())
    }
}

func (ks *KeyStore) get(name string) *Key {
    var coll = ks.session.Copy().DB(config.mongoDb).C(config.mongoColl)

    key := &Key{}
    err := coll.Find(bson.M{"name": name}).One(key)
    if err != nil {
        fmt.Println("Error get key: ", err.Error())
    }

    return key
}
