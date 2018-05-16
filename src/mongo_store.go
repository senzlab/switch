package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

type Key struct {
	Name     string
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
		log.Printf("Error put key: ", err.Error(), ": "+key.Name)
	}
}

func (ks *MongoStore) getKey(name string) *Key {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.keyColl)
	key := &Key{}
	err := coll.Find(bson.M{"name": name}).One(key)
	if err != nil {
		fmt.Printf("key not found: ", err.Error(), ": "+name)
	}

	return key
}

func (ks *MongoStore) enqueueSenz(qSenz *Senz) {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.senzColl)
	err := coll.Insert(qSenz)
	if err != nil {
		log.Printf("Error enque senz: ", err.Error())
	}
}

func (ks *MongoStore) dequeueSenzById(uid string) *Senz {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.senzColl)

	// get
	qSenz := &Senz{}
	gErr := coll.Find(bson.M{"uid": uid}).One(qSenz)
	if gErr != nil {
		log.Printf("No deque senz uid: ", uid)
	}

	// then remove
	rErr := coll.Remove(bson.M{"uid": uid})
	if rErr != nil {
		log.Printf("No remove senz uid: ", uid)
	}

	return qSenz
}

func (ks *MongoStore) dequeueSenzByReceiver(receiver string) []Senz {
	sessionCopy := ks.session.Copy()
	defer sessionCopy.Close()

	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.senzColl)

	log.Printf("dequeuing senz of : ", receiver)

	// get
	var qSenzes []Senz
	gErr := coll.Find(bson.M{"receiver": receiver}).All(&qSenzes)
	if gErr != nil {
		log.Printf("Error deque senz: ", gErr.Error())
	}

	// senz id to delete
	var dSenzes []string
	for _, z := range qSenzes {
		dSenzes = append(dSenzes, z.Uid)
	}

	// then remove all
	_, rErr := coll.RemoveAll(bson.M{"uid": bson.M{"$in": dSenzes}})
	if rErr != nil {
		log.Printf("Error remove key: ", rErr.Error())
	}

	return qSenzes
}
