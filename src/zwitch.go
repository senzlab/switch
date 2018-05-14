package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"io"
	"io/ioutil"
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
}

type SenzMsg struct {
	Uid string
	Msg string
}

// global
// 1. mongo store
var (
	mongoStore = &MongoStore{}
)

func main() {
	// db setup
	session, err := mgo.Dial(mongoConfig.mongoHost)
	if err != nil {
		fmt.Println("Error connecting mongo: ", err.Error())
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	mongoStore.session = session

	// init key pair
	setUpKeys()

	// router
	r := mux.NewRouter()
	r.HandleFunc("/promizes", postPromize).Methods("POST")
	r.HandleFunc("/blobs", getBlob).Methods("POST")
	r.HandleFunc("/uzers", postUzer).Methods("POST")
	r.HandleFunc("/uzers", putUzer).Methods("PUT")
	r.HandleFunc("/connections", postConnection).Methods("POST")

	// start server
	err = http.ListenAndServe(":7171", r)
	if err != nil {
		println(err.Error)
		os.Exit(1)
	}
}

func postPromize(w http.ResponseWriter, r *http.Request) {
	println("posging promize")
	// read body
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	println(string(b))

	// unmarshel json and parse senz
	var senzMsg SenzMsg
	json.Unmarshal(b, &senzMsg)
	senz, err := parse(senzMsg.Msg)
	if err != nil {
		// we not send any response we just disconnect
		errorResponse(w, "", "")
		return
	}

	// get senzie key
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	senzieKey := getSenzieRsaPub(mongoStore.getKey(senz.Sender).Value)
	err = verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}

	// promize and get response
	b, statusCode := post(senz)
	if statusCode != http.StatusOK {
		errorResponse(w, senz.Attr["uid"], senz.Sender)
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
			successResponse(w, z.Attr["uid"], z.Receiver)
		} else {
			// this means forwarding promize
			// enqueu promizes
			mongoStore.enqueueSenz(z)

			// check receiver exists
			// TODO
			// get device id from mongo store key
			// send push notification to reciver
			rKey := mongoStore.getKey(z.Receiver)
			to := rKey.DeviceId
			senzMsg := SenzMsg{
				Uid: z.Attr["uid"],
				Msg: notifyPromizeSenz(z),
			}
			notify(to, senzMsg)
		}
	}

	return
}

func getBlob(w http.ResponseWriter, r *http.Request) {
	println("getting blob")

	// read body
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	println(string(b))

	// unmarshel json
	var senzMsg SenzMsg
	json.Unmarshal(b, &senzMsg)
	senz, err := parse(senzMsg.Msg)
	if err != nil {
		// we jsut retun with out sending response
		errorResponse(w, "", "")
		return
	}

	// get senzie key
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	senzieKey := getSenzieRsaPub(mongoStore.getKey(senz.Sender).Value)
	err = verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}

	// get senz
	qSenz := mongoStore.dequeueSenzById(senz.Attr["uid"])
	if qSenz.Receiver != senz.Sender {
		// not authorized
		errorResponse(w, "", "")
		return
	}

	// response blob
	blobResponse(w, qSenz.Attr["blob"], senz.Attr["uid"], senz.Sender)
	return
}

func postUzer(w http.ResponseWriter, r *http.Request) {
	// read body
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	println(string(b))

	// unmarshel json
	var senzMsg SenzMsg
	json.Unmarshal(b, &senzMsg)
	senz, err := parse(senzMsg.Msg)
	if err != nil {
		// we jsut retun with out sending response
		errorResponse(w, "", "")
		return
	}

	// check weather user exists
	key := mongoStore.getKey(senz.Sender)
	if key.Value == "" {
		// this means no senzie
		// verify signature
		payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
		senzieKey := getSenzieRsaPub(senz.Attr["pubkey"])
		err = verify(payload, senz.Digsig, senzieKey)
		if err != nil {
			errorResponse(w, senz.Attr["uid"], senz.Sender)
			return
		}

		// post user to chainz
		// handle response
		_, statusCode := post(senz)
		if statusCode != http.StatusOK {
			errorResponse(w, senz.Attr["uid"], senz.Sender)
			return
		}

		// save user with
		// 1. key
		// 2. firebase device id
		key := Key{
			Name:     senz.Sender,
			Value:    senz.Attr["pubkey"],
			DeviceId: senz.Attr["did"],
		}
		mongoStore.putKey(&key)

		// success response
		successResponse(w, senz.Attr["uid"], senz.Sender)
		return
	} else {
		// this means already registered senzie
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}
}

func putUzer(w http.ResponseWriter, r *http.Request) {
	// read body
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	println(string(b))

	// unmarshel json
	var senzMsg SenzMsg
	json.Unmarshal(b, &senzMsg)
	senz, err := parse(senzMsg.Msg)
	if err != nil {
		// we jsut retun with out sending response
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}
	println(senz)

	// get senzie key
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	senzieKey := getSenzieRsaPub(mongoStore.getKey(senz.Sender).Value)
	err = verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}

	// post user to chainz
	// handle response
	_, statusCode := post(senz)
	if statusCode != http.StatusOK {
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}

	// success response
	successResponse(w, senz.Attr["uid"], senz.Sender)
	return
}

func postConnection(w http.ResponseWriter, r *http.Request) {
	// read body
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	println(string(b))

	// unmarshel json
	var senzMsg SenzMsg
	json.Unmarshal(b, &senzMsg)
	senz, err := parse(senzMsg.Msg)
	if err != nil {
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}
	println(senz)

	// get senzie key
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	senzieKey := getSenzieRsaPub(mongoStore.getKey(senz.Sender).Value)
	err = verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}

	// check receiver exists
	rKey := mongoStore.getKey(senz.Receiver)
	if rKey.Value == "" {
		// no reciver exists
		errorResponse(w, senz.Attr["uid"], senz.Sender)
		return
	}

	// get device id from mongo store key
	// send push notification to reciver
	to := rKey.DeviceId
	senzMsg = SenzMsg{
		Uid: senz.Attr["uid"],
		Msg: notifyConnectSenz(senz),
	}
	notify(to, senzMsg)

	// success response
	successResponse(w, senz.Attr["uid"], senz.Sender)
	return
}

func errorResponse(w http.ResponseWriter, uid string, to string) {
	// marshel and return error
	zmsg := SenzMsg{
		Uid: uid,
		Msg: statusSenz("ERROR", uid, to),
	}
	j, _ := json.Marshal(zmsg)
	http.Error(w, string(j), 400)
}

func successResponse(w http.ResponseWriter, uid string, to string) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	zmsg := SenzMsg{
		Uid: uid,
		Msg: statusSenz("SUCCESS", uid, to),
	}
	j, _ := json.Marshal(zmsg)
	io.WriteString(w, string(j))
}

func blobResponse(w http.ResponseWriter, blob string, uid string, to string) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	zmsg := SenzMsg{
		Uid: uid,
		Msg: blobSenz(blob, uid, to),
	}
	j, _ := json.Marshal(zmsg)
	io.WriteString(w, string(j))
}
