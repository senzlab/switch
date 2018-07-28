package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Senz struct {
	Msg      string
	Uid      string
	Ztype    string
	Sender   string
	Receiver string
	Attr     map[string]string
	Digsig   string
	Status   string
}

type SenzMsg struct {
	Uid string
	Msg string
}

// mongo store
var mongoStore = &MongoStore{}

func main() {
	// setup logging
	initLogz()

	// db setup
	info := &mgo.DialInfo{
		Addrs:    []string{mongoConfig.mongoHost},
		Database: mongoConfig.mongoDb,
		Username: mongoConfig.username,
		Password: mongoConfig.password,
	}
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Printf("ERROR connecting mongo, %s ", err.Error())
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	mongoStore.session = session

	// init key pair
	setUpKeys()

	// listen for incoming conns
	initHttpz()
}

func initHttpz() {
	// router
	r := mux.NewRouter()
	r.HandleFunc("/igift/v1/contractz", contractz).Methods("POST")

	// start server
	err := http.ListenAndServe(":7171", r)
	if err != nil {
		log.Printf("ERROR init httpz, ", err.Error())
		os.Exit(1)
	}
}

func contractz(w http.ResponseWriter, r *http.Request) {
	// read body
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	log.Printf("new contract senz, %s", string(b))

	// unmarshel json and parse senz
	var zmsg SenzMsg
	json.Unmarshal(b, &zmsg)
	senz, err := parse(zmsg.Msg)
	if err != nil {
		log.Printf("ERROR senz, %s", err.Error())

		// error response
		return
	}

	if senz.Receiver == config.switchName {
		// this could be
		// 1. reg senz
		// 2. fetch senz
		// 3. connect senz
		if senz.Ztype == "PUT" {
			handleReg(w, senz)
			return
		}

		if senz.Ztype == "GET" {
			// this is fetch
			handleFetch(w, senz)
			return
		}

		if senz.Ztype == "SHARE" {
			// this is connect
			handleConnect(w, senz)
			return
		}
	}

	if senz.Receiver == chainzConfig.name {
		// this if for chainz
		handlePromize(w, senz)
		return
	}
}

func handleReg(w http.ResponseWriter, senz *Senz) {
	// this is reg
	// check weather user exists
	key := mongoStore.getKey(senz.Sender)
	if key.Value == "" {
		// this means no senzie
		// verify signature
		payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
		senzieKey := getSenzieRsaPub(senz.Attr["pubkey"])
		err := verify(payload, senz.Digsig, senzieKey)
		if err != nil {
			return
		}

		// post user to chainz
		// handle response
		_, statusCode := post(senz)
		if statusCode != http.StatusOK {
			// error from bankz
			zmsg := SenzMsg{
				Uid: senz.Attr["uid"],
				Msg: statusSenz("400", senz.Attr["uid"], senz.Sender),
			}
			responze(w, []SenzMsg{zmsg})
			return
		}

		// save user with
		// 1. key
		// 2. firebase/apn device id
		key := Key{
			Name:     senz.Sender,
			Password: "lambda",
			Value:    senz.Attr["pubkey"],
			Device:   senz.Attr["dev"],
			DeviceId: senz.Attr["devid"],
		}
		mongoStore.putKey(&key)

		// status response
		zmsg := SenzMsg{
			Uid: senz.Attr["uid"],
			Msg: statusSenz("200", senz.Attr["uid"], senz.Sender),
		}
		responze(w, []SenzMsg{zmsg})

		return
	} else {
		// this means already registered senzie
		zmsg := SenzMsg{
			Uid: senz.Attr["uid"],
			Msg: statusSenz("403", senz.Attr["uid"], senz.Sender),
		}
		responze(w, []SenzMsg{zmsg})
		return
	}
}

func handleConnect(w http.ResponseWriter, senz *Senz) {
	// verify senz first
	err := verifySenz(senz)
	if err != nil {
		return
	}

	// check receiver exists
	rKey := mongoStore.getKey(senz.Attr["to"])
	if rKey.Value == "" {
		// no reciver exists
		// error response
		zmsg := SenzMsg{
			Uid: senz.Attr["uid"],
			Msg: statusSenz("404", senz.Attr["uid"], senz.Sender),
		}
		responze(w, []SenzMsg{zmsg})
		return
	}

	// get device id from mongo store key
	// send push notification to reciver
	to := rKey.DeviceId
	nz := notifyConnectSenz(senz)
	if rKey.Device == "android" {
		notifa(to, AndroidNotification{nz})
	} else {
		notifi(to, AppleNotification{"New contact", "iGift contact request received", nz})
	}

	// success response
	zmsg := SenzMsg{
		Uid: senz.Attr["uid"],
		Msg: statusSenz("200", senz.Attr["uid"], senz.Sender),
	}
	responze(w, []SenzMsg{zmsg})
}

func handlePromize(w http.ResponseWriter, senz *Senz) {
	// verify senz first
	err := verifySenz(senz)
	if err != nil {
		return
	}

	// post promize for chainz
	b, statusCode := post(senz)
	if statusCode != http.StatusOK {
		zmsg := SenzMsg{
			Uid: senz.Attr["uid"],
			Msg: statusSenz("400", senz.Attr["uid"], senz.Sender),
		}
		responze(w, []SenzMsg{zmsg})
		return
	}

	var zmsgs []SenzMsg
	json.Unmarshal(b, &zmsgs)

	// iterate over each and every msg and process
	for _, zmsg := range zmsgs {
		z, _ := parse(string(zmsg.Msg))

		if z.Receiver == senz.Sender {
			// this message for senz sender
			// send success response back
			zmsg := SenzMsg{
				Uid: senz.Attr["uid"],
				Msg: statusSenz("200", senz.Attr["uid"], senz.Sender),
			}
			responze(w, []SenzMsg{zmsg})
		} else {
			// this means forwarding promize
			// enqueu promizes
			z.Status = "0"
			mongoStore.enqueueSenz(z)

			// get device id from mongo store key
			// send push notification to reciver
			rKey := mongoStore.getKey(z.Receiver)
			to := rKey.DeviceId
			nz := notifyPromizeSenz(z)
			if rKey.Device == "android" {
				notifa(to, AndroidNotification{nz})
			} else {
				notifi(to, AppleNotification{"New iGift", "New iGift received", nz})
			}
		}
	}
}

func handleFetch(w http.ResponseWriter, senz *Senz) {
	// verify senz first
	err := verifySenz(senz)
	if err != nil {
		return
	}

	// check for version
	if _, ok := senz.Attr["version"]; ok {
		// response version
		zmsg := SenzMsg{
			Uid: senz.Attr["uid"],
			Msg: versionSenz(senz.Attr["uid"], senz.Sender),
		}
		var zmsgs []SenzMsg
		zmsgs = append(zmsgs, zmsg)
		responze(w, zmsgs)
		return
	}

	// check for senzes
	if _, ok := senz.Attr["senzes"]; ok {
		// get all unfetched senzes
		qSenzes := mongoStore.dequeueSenzByReceiver(senz.Sender)
		var zmsgs []SenzMsg
		for _, z := range qSenzes {
			// append qsenz
			zmsg := SenzMsg{
				Uid: z.Attr["uid"],
				Msg: metaSenz(z, senz.Sender),
			}
			zmsgs = append(zmsgs, zmsg)
		}
		responze(w, zmsgs)
	} else {
		// response version
		// get senz
		qSenz := mongoStore.dequeueSenzById(senz.Attr["uid"])
		if qSenz.Receiver != senz.Sender {
			// not authorized
			log.Printf("ERROR: not authorized to get blob")
			return
		}

		// response blob
		zmsg := SenzMsg{
			Uid: qSenz.Attr["uid"],
			Msg: blobSenz(qSenz.Attr["blob"], qSenz.Attr["uid"], senz.Sender),
		}
		var zmsgs []SenzMsg
		zmsgs = append(zmsgs, zmsg)
		responze(w, zmsgs)
	}
}

func verifySenz(senz *Senz) error {
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	key := mongoStore.getKey(senz.Sender)
	if key.Value == "" {
		log.Printf("ERROR: cannot verify signarue, no key found")
		return errors.New("No senzie key found")
	}

	senzieKey := getSenzieRsaPub(key.Value)
	err := verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		log.Printf("ERROR: cannot verify signarue, so dorp the conneciton")
		return errors.New("Cannot verify signature")
	}

	return nil
}

func responze(w http.ResponseWriter, zmsgs []SenzMsg) {
	j, _ := json.Marshal(zmsgs)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(j))
}
