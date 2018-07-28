package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

type Key struct {
	Name     string
	Password string
	Value    string
	Device   string
	DeviceId string
}

type MongoStore struct {
	session *mgo.Session
}

func (ks *MongoStore) putKey(key *Key) {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.keyColl)
	err := coll.Insert(key)
	if err != nil {
		log.Printf("ERROR: put key fail, %s, keyName: %s", err.Error(), key.Name)
	}
}

func (ks *MongoStore) getKey(name string) *Key {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.keyColl)
	key := &Key{}
	err := coll.Find(bson.M{"name": name}).One(key)
	if err != nil {
		log.Printf("ERROR: key not found, %s, name: %s", err.Error(), name)
	}

	return key
}

func (ks *MongoStore) enqueueSenz(qSenz *Senz) {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.senzColl)
	err := coll.Insert(qSenz)
	if err != nil {
		log.Printf("ERROR: fail enque senz, %s", err.Error())
	}
}

func (ks *MongoStore) dequeueSenzById(uid string) *Senz {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.senzColl)

	// get
	qSenz := &Senz{}
	err := coll.Find(bson.M{"uid": uid}).One(qSenz)
	if err != nil {
		log.Printf("ERROR: no deque senz uid, %s", uid)
	}

	// then update delivery status
	err = coll.Update(bson.M{"uid": uid}, bson.M{"$set": bson.M{"status": "1"}})
	if err != nil {
		log.Printf("ERROR: fail update delivevery state uid, %s", uid)
	}

	return qSenz
}

func (ks *MongoStore) dequeueSenzByReceiver(receiver string) []Senz {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.senzColl)

	// get
	var qSenzes []Senz
	err := coll.Find(bson.M{"$and": []bson.M{{"receiver": receiver}, {"status": "0"}}}).
		Select(bson.M{"uid": 1, "attr": 1}).All(&qSenzes)
	if err != nil {
		log.Printf("ERROR: fail deque senz, %s", err.Error())
	}

	return qSenzes
}
