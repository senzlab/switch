package main

import (
	"bufio"
	"fmt"
	"gopkg.in/mgo.v2"
	"net"
	"strings"
	"time"
)

type Senzie struct {
	name   string
	id     string
	out    chan Senz
	quit   chan bool
	tik    *time.Ticker
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

// constants
// 1. buffer size
// 2. socket read timeout
// 3. ticking interval
const (
	chanelSize  = 10
	bufSize     = 16 * 1024
	readTimeout = 30 * time.Minute
	tikInterval = 60 * time.Second
)

// global
// 1. connected senzies
// 2. mongo store
var (
	senzies    = map[string]*Senzie{}
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

	// listen for incoming conns
	listener, err := net.Listen("tcp", ":"+config.switchPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	// listeneing
	listening(listener)
}

func listening(listener net.Listener) {
LISTENER:
	for {
		// handle new connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue LISTENER
		}

		// new senzie
		senzie := &Senzie{
			out:    make(chan Senz, chanelSize),
			quit:   make(chan bool),
			reader: bufio.NewReaderSize(conn, bufSize),
			writer: bufio.NewWriterSize(conn, bufSize),
			conn:   conn,
		}

		go registering(senzie)
	}
}

func registering(senzie *Senzie) {
REGISTER:
	for {
		// listen for reg senz
		msg, err := senzie.reader.ReadString(';')
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			senzie.conn.Close()
			break REGISTER
		}

		println("received " + msg)

		// not handle TAK, TIK, TUK
		if msg == "TAK;" || msg == "TIK;" || msg == "TUK;" {
			continue REGISTER
		}

		senz, err := parse(msg)
		if err != nil {
			fmt.Println("Error senz: ", err.Error())
			senzie.conn.Close()
			break REGISTER
		}

		senzie.name = senz.Sender
		senzie.id = senz.Attr["uid"]

		// get pubkey
		pubkey := senz.Attr["pubkey"]
		key := mongoStore.getKey(senzie.name)

		// check for reg
		if key.Value == "" {
			// not registerd senzie
			// save pubkey
			// add senzie
			mongoStore.putKey(&Key{senzie.name, pubkey})
			senzies[senzie.name] = senzie

			// send status
			sz := regSenz(senz.Attr["uid"], "REG_DONE", senzie.name)
			senzie.writer.WriteString(sz.Msg + ";")
			senzie.writer.Flush()

			// start ticking
			// start reading
			// start writing
			senzie.tik = time.NewTicker(tikInterval)
			go reading(senzie)
			go writing(senzie)

			break REGISTER
		} else if key.Value == pubkey {
			// already registerd senzie
			// close existing senzie's conn
			// then add new senzie
			if senzies[senzie.name] != nil {
				senzies[senzie.name].conn.Close()
			}
			senzies[senzie.name] = senzie

			// send status
			sz := regSenz(senz.Attr["uid"], "REG_ALR", senzie.name)
			senzie.writer.WriteString(sz.Msg + ";")
			senzie.writer.Flush()

			// start ticking
			// start reading
			// start writing
			// dispatch queued messages of senzie
			senzie.tik = time.NewTicker(tikInterval)
			go reading(senzie)
			go writing(senzie)
			go dispatching(senzie)

			break REGISTER
		} else {
			// name already obtained
			// send status
			uid := senz.Attr["uid"]
			senz := regSenz(uid, "REG_FAIL", senzie.name)

			// write
			senzie.writer.WriteString(senz.Msg + ";")
			senzie.writer.Flush()

			break REGISTER
		}
	}
}

func reading(senzie *Senzie) {
READER:
	for {
		msg, err := senzie.reader.ReadString(';')
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			break READER
		}

		// set read deadline to detect dead peers
		senzie.conn.SetReadDeadline(time.Now().Add(readTimeout))

		// not handle TAK, TIK, TUK
		if msg == "TAK;" || msg == "TIK;" || msg == "TUK;" {
			continue READER
		}

		// parse senz and handle it
		senz, err := parse(msg)
		if err != nil {
			fmt.Println("Error senz: ", err.Error())
			senzie.conn.Close()
			break READER
		}

		println("received: " + msg)

		// verify signature first of all
		payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
		senzieKey := getSenzieRsaPub(mongoStore.getKey(senzie.name).Value)
		err = verify(payload, senz.Digsig, senzieKey)
		if err != nil {
			println("cannot verify signarue, so dorp the conneciton")
			break READER
		}

		if senz.Receiver == config.switchName {
			if senz.Ztype == "GET" {
				// this is requesting pub key of other senzie
				// fing pubkey and send
				key := mongoStore.getKey(senz.Attr["name"])
				uid := senz.Attr["uid"]
				name := senz.Attr["name"]
				senzie.out <- keySenz(uid, key.Value, name, senzie.name)
			} else if senz.Ztype == "AWA" {
				// this means message delivered to senzie
				// get senz with given uid
				uid := senz.Attr["uid"]
				var dz = mongoStore.dequeueSenzById(uid)
				if dz.Ztype == "" || dz.Sender == config.switchName || dz.Sender == chainzConfig.name {
					continue READER
				}

				// find sender and send GIYA
				gz := giyaSenz(uid, dz.Sender)
				if senzies[dz.Sender] != nil {
					senzies[dz.Sender].out <- gz
				} else {
					fmt.Println("no senzie to send giya senz, enqueue " + gz.Msg)
					mongoStore.enqueueSenz(&gz)
				}
			}
		} else if senz.Receiver == chainzConfig.name {
			// for sampath bank
			go promize(senz)
		} else {
			// send AWA back to sender
			uid := senz.Attr["uid"]
			senzie.out <- awaSenz(uid, senzie.name)

			// forwared senz msg to receiver
			if senzies[senz.Receiver] != nil {
				writeRecover(senzies[senz.Receiver], senz)
			} else {
				fmt.Println("no senzie to send senz, enqueued " + senz.Msg)
				mongoStore.enqueueSenz(senz)
			}
		}
	}

	println("exit reader...")

	// quit all routeins of this senzie
	// close conn
	senzie.quit <- true
	senzie.conn.Close()
}

func writing(senzie *Senzie) {
WRITER:
	for {
		select {
		case <-senzie.quit:
			println("quiting/write/tick -- " + senzie.name)

			// close channles
			senzie.tik.Stop()
			close(senzie.quit)
			close(senzie.out)

			// delete senzie
			// check weather deleting senzie is same to senzie in senzies
			if senzie.id == senzies[senzie.name].id {
				delete(senzies, senzie.name)
			}
			break WRITER
		case senz := <-senzie.out:
			// enqueu senz, except AWA, GIYA and broadcase senz(receiver="*")
			if senz.Ztype != "AWA" && senz.Ztype != "GIYA" {
				mongoStore.enqueueSenz(&senz)
			}

			senzie.writer.WriteString(senz.Msg + ";")
			senzie.writer.Flush()
		case <-senzie.tik.C:
			println("ticking -- " + senzie.name)
			senzie.writer.WriteString("TIK;")
			senzie.writer.Flush()
		}
	}
}

func dispatching(senzie *Senzie) {
	// find queued messages from mongo store
	var zs = mongoStore.dequeueSenzByReceiver(senzie.name)

	fmt.Println("despatching ... ", len(zs))

	// dispatch queued messages to senzie
	for _, z := range zs {
		senzie.out <- z
	}
}

func writeRecover(senzie *Senzie, z *Senz) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// TODO enqueu senz

	// write
	senzie.out <- *z
}
