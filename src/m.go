package main

import (
	"gopkg.in/mgo.v2"
	"log"
)

func m() {
	info := &mgo.DialInfo{
		Addrs:    []string{mongoConfig.mongoHost},
		Database: mongoConfig.mongoDb,
		Username: mongoConfig.username,
		Password: mongoConfig.password,
	}

	// db setup
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Printf("Error connecting mongo: ", err.Error())
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	mongoStore.session = session
	println("done...")

	k := mongoStore.getKey("+94775432015")
	println(k.Name)
	println(k.Password)
}
