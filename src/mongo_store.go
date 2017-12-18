package main

import (
    "fmt"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

type Key struct {
    Name        string
    Value       string
}

type QSenz struct {
    uid         string
    msg         string
    sender      string
    receiver    string 
}

type MongoStore struct {
    session *mgo.Session
}

func (ks *MongoStore) put(key *Key) {
    sessionCopy := ks.session.Copy() 
    defer sessionCopy.Close()

    var coll = sessionCopy.DB(config.mongoDb).C(config.keyColl)
    err := coll.Insert(key)
    if err != nil {
        fmt.Println("Error put key: ", err.Error())
    }
}

func (ks *MongoStore) get(name string) *Key {
    sessionCopy := ks.session.Copy() 
    defer sessionCopy.Close()

    var coll = sessionCopy.DB(config.mongoDb).C(config.keyColl)
    key := &Key{}
    err := coll.Find(bson.M{"name": name}).One(key)
    if err != nil {
        fmt.Println("Error get key: ", err.Error())
    }

    return key
}

func (ks *MongoStore) enqueueSenz(qSenz *QSenz) {
    sessionCopy := ks.session.Copy() 
    defer sessionCopy.Close()

    var coll = sessionCopy.DB(config.mongoDb).C(config.senzColl)
    err := coll.Insert(qSenz)
    if err != nil {
        fmt.Println("Error put key: ", err.Error())
    }
}

func (ks *MongoStore) dequeueSenzById(uid string) *QSenz {
    sessionCopy := ks.session.Copy() 
    defer sessionCopy.Close()

    var coll = sessionCopy.DB(config.mongoDb).C(config.senzColl)

    // get
    qSenz := &QSenz{}
    gErr := coll.Find(bson.M{"uid": uid}).One(qSenz)
    if gErr != nil {
        fmt.Println("Error get key: ", gErr.Error())
    }

    // then remove
    rErr := coll.Remove(bson.M{"uid": "uid"})
	if rErr != nil {
        fmt.Println("Error remove key: ", rErr.Error())
	}

    return qSenz
}

func (ks *MongoStore) dequeueSenzByReceiver(receiver string) []QSenz {
    sessionCopy := ks.session.Copy()
    defer sessionCopy.Close()

    var coll = sessionCopy.DB(config.mongoDb).C(config.senzColl)

    // get
    var qSenzes []QSenz
    gErr := coll.Find(bson.M{"receiver": receiver}).All(qSenzes)
    if gErr != nil {
        fmt.Println("Error get key: ", gErr.Error())
    }

    // then remove
    _, rErr := coll.RemoveAll(bson.M{"receiver": receiver})
	if rErr != nil {
        fmt.Println("Error remove key: ", rErr.Error())
	}

    return qSenzes
}
