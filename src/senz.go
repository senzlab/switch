package main

import (
	"bufio"
	"encoding/json"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"gopkg.in/mgo.v2"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type Senzie struct {
	reader *bufio.Reader
	writer *bufio.Writer
	conn   net.Conn
}

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

// constants
// 1. buffer size
// 2. socket read timeout
const (
	bufSize     = 16 * 1024
	readTimeout = 30 * time.Minute
)

// global
// 1. connected senzies
// 2. mongo store
// 3. apn client
var (
	senzies    = map[string]*Senzie{}
	mongoStore = &MongoStore{}
	apnClient  *apns2.Client
)

func main() {
	// db setup
	session, err := mgo.Dial(mongoConfig.mongoHost)
	if err != nil {
		log.Printf("Error connecting mongo: ", err.Error())
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	mongoStore.session = session

	// init key pair
	setUpKeys()

	// setup apn push setup
	apnCert, err := certificate.FromP12File(apnConfig.certificate, "")
	if err != nil {
		log.Printf("APN cert load Error:", err)
	}
	apnClient = apns2.NewClient(apnCert).Development()

	// listen for incoming conns
	listener, err := net.Listen("tcp", ":"+config.switchPort)
	if err != nil {
		log.Printf("Error listening:", err.Error())
		return
	}
	defer listener.Close()
	listening(listener)
}

func listening(listener net.Listener) {
LISTENER:
	// listeneing
	for {
		// handle new connections
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting: ", err.Error())
			continue LISTENER
		}

		// new senzie
		senzie := &Senzie{
			reader: bufio.NewReaderSize(conn, bufSize),
			writer: bufio.NewWriterSize(conn, bufSize),
			conn:   conn,
		}

		go reading(senzie)
	}
}

func reading(senzie *Senzie) {
	msg, err := senzie.reader.ReadString(';')
	if err != nil {
		log.Printf("Error reading: ", err.Error())
		senzie.conn.Close()
		return
	}

	// parse senz and handle it
	senz, err := parse(msg)
	if err != nil {
		log.Printf("Error senz: ", err.Error())
		senzie.conn.Close()
		return
	}

	log.Printf("received senz: ", msg)

	if senz.Receiver == config.switchName {
		// this could be
		// 1. reg senz
		// 2. fetch senz
		// 3. connect senz
		if senz.Ztype == "PUT" {
			handleReg(senzie, senz)
			senzie.conn.Close()
			return
		}

		if senz.Ztype == "GET" {
			// this is fetch
			handleFetch(senzie, senz)
			senzie.conn.Close()
			return
		}

		if senz.Ztype == "SHARE" {
			// this is connect
			handleConnect(senzie, senz)
			senzie.conn.Close()
			return
		}
	}

	if senz.Receiver == chainzConfig.name {
		// this if for chainz
		handlePromize(senzie, senz)
		senzie.conn.Close()
		return
	}

	// heare means invalid senzes
	senzie.conn.Close()
}

func handleReg(senzie *Senzie, senz *Senz) {
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
			return
		}

		// save user with
		// 1. key
		// 2. firebase/apn device id
		key := Key{
			Name:     senz.Sender,
			Value:    senz.Attr["pubkey"],
			Device:   senz.Attr["dev"],
			DeviceId: senz.Attr["devid"],
		}
		mongoStore.putKey(&key)

		// success response
		sz := statusSenz("SUCCESS", senz.Attr["uid"], senz.Sender)
		senzie.writer.WriteString(sz + ";")
		senzie.writer.Flush()

		return
	} else {
		// this means already registered senzie
		return
	}
}

func handleFetch(senzie *Senzie, senz *Senz) {
	// verify signature first of all
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	senzieKey := getSenzieRsaPub(mongoStore.getKey(senz.Sender).Value)
	err := verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		log.Printf("cannot verify signarue, so dorp the conneciton")
		return
	}

	// check receiver exists
	rKey := mongoStore.getKey(senz.Receiver)
	if rKey.Value == "" {
		// no reciver exists
		log.Printf("not receiver: ", senz.Receiver)
		return
	}

	// get senz
	qSenz := mongoStore.dequeueSenzById(senz.Attr["uid"])
	if qSenz.Receiver != senz.Sender {
		// not authorized
		log.Printf("not authorized to get blob")
		return
	}

	// response blob
	bz := blobSenz(qSenz.Attr["blob"], qSenz.Attr["uid"], senz.Sender)
	senzie.writer.WriteString(bz + ";")
	senzie.writer.Flush()

	return
}

func handleConnect(senzie *Senzie, senz *Senz) {
	// verify signature first of all
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	senzieKey := getSenzieRsaPub(mongoStore.getKey(senz.Sender).Value)
	err := verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		log.Printf("cannot verify signarue, so dorp the conneciton")
		return
	}

	// check receiver exists
	rKey := mongoStore.getKey(senz.Attr["to"])
	if rKey.Value == "" {
		// no reciver exists
		return
	}

	// get device id from mongo store key
	// send push notification to reciver
	to := rKey.DeviceId
	nz := notifyConnectSenz(senz)
	if rKey.Device == "android" {
		notifa(to, AndroidNotification{nz})
	} else {
		notifi(apnClient, to, AppleNotification{"New contact", "senz_connect", nz})
	}

	// success response
	sz := statusSenz("SUCCESS", senz.Attr["uid"], senz.Sender)
	senzie.writer.WriteString(sz + ";")
	senzie.writer.Flush()
}

func handlePromize(senzie *Senzie, senz *Senz) {
	// verify signature first of all
	payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
	senzieKey := getSenzieRsaPub(mongoStore.getKey(senz.Sender).Value)
	err := verify(payload, senz.Digsig, senzieKey)
	if err != nil {
		log.Printf("cannot verify signarue, so dorp the conneciton")
		return
	}

	// post promize for chainz
	b, statusCode := post(senz)
	if statusCode != http.StatusOK {
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
			sz := statusSenz("SUCCESS", senz.Attr["uid"], senz.Sender)
			senzie.writer.WriteString(sz + ";")
			senzie.writer.Flush()
		} else {
			// this means forwarding promize
			// enqueu promizes
			mongoStore.enqueueSenz(z)

			// get device id from mongo store key
			// send push notification to reciver
			rKey := mongoStore.getKey(z.Receiver)
			to := rKey.DeviceId
			nz := notifyPromizeSenz(z)
			if rKey.Device == "android" {
				notifa(to, AndroidNotification{nz})
			} else {
				notifi(apnClient, to, AppleNotification{"New iGift", "senz_igift", nz})
			}
		}
	}
}
