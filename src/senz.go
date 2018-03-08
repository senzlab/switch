package main

import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
    "time"
    "gopkg.in/mgo.v2"
)

type Senzie struct {
    name        string
	out         chan Senz
    quit        chan bool
    tik         *time.Ticker
    reader      *bufio.Reader
    writer      *bufio.Writer
    conn        net.Conn
}

type Senz struct {
    Msg         string
    Uid         string
    Ztype       string
    Sender      string
    Receiver    string
    Attr        map[string]string
    Digsig      string
}

// constants
// 1. buffer size
// 2. socket read timeout
// 3. ticking interval
const bufSize = 16 * 1024
const readTimeout = 5 * time.Minute
const tikInterval = 60 * time.Second

// keep connected senzies
var senzies = map[string]*Senzie{}
var mongoStore = &MongoStore{}

func main() {
    // first init key pair
    setUpKeys()

    // listen for incoming conns
    l, err := net.Listen("tcp", ":" + config.switchPort)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }

    // close listern on app closes
    defer l.Close()

    fmt.Println("Listening on " + config.switchPort)

    // db setup
    session, err:= mgo.Dial(config.mongoHost)
    if err != nil {
        fmt.Println("Error connecting mongo: ", err.Error())
        os.Exit(1)
    }

    // close session on app closes
    defer session.Close()

    session.SetMode(mgo.Monotonic, true)
    mongoStore.session = session

    LISTENER:
    for {
        // handle new connections
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            continue LISTENER
        }

        // enable keep alive
	    //conn.(*net.TCPConn).SetKeepAlive(true)

        // new senzie
        senzie := &Senzie {
            out: make(chan Senz),
            quit: make(chan bool),
            reader: bufio.NewReaderSize(conn, bufSize),
            writer: bufio.NewWriterSize(conn, bufSize),
            conn: conn,
        }

        go registering(senzie)
    }
}

func registering(senzie *Senzie) {
    // listen for reg senz
    msg, err := senzie.reader.ReadString(';')
    if err != nil {
        fmt.Println("Error reading: ", err.Error())

        senzie.conn.Close()
    }

    println("received " + msg)

    // get pubkey
    senz := parse(msg)
    senzie.name = senz.Sender
    pubkey := senz.Attr["pubkey"]
    key := mongoStore.getKey(senzie.name)

    // check for reg
    if(key.Value == "") {
        // not registerd senzie
        // save pubkey
        // add senzie
        mongoStore.putKey(&Key{senzie.name, pubkey})
        senzies[senzie.name] = senzie

        // start ticking
        // start reading
        // start writing
        senzie.tik = time.NewTicker(tikInterval)
        go reading(senzie)
        go writing(senzie)

        // send status
        uid := senz.Attr["uid"]
        senzie.out <- regSenz(uid, "REG_DONE", senzie.name)
    } else if(key.Value == pubkey) {
        // already registerd senzie
        // close existing senzie's conn
        // delete existing senzie
        // then add new senzie
        if senzies[senzie.name] != nil {
            senzies[senzie.name].conn.Close()
            delete(senzies, senzie.name)
        }
        senzies[senzie.name] = senzie

        // start ticking
        // start reading
        // start writing
        senzie.tik = time.NewTicker(tikInterval)
        go reading(senzie)
        go writing(senzie)

        // send status
        uid := senz.Attr["uid"]
        senzie.out <- regSenz(uid, "REG_ALR", senzie.name)

        // dispatch queued messages of senzie
        go dispatching(senzie)
    } else {
        // name already obtained
        // send status
        uid := senz.Attr["uid"]
        senz := regSenz(uid, "REG_FAIL", senzie.name)

        // write
        senzie.writer.WriteString(senz.Msg + ";")
        senzie.writer.Flush()
    }
}

func reading(senzie *Senzie) {
    // read senz
    READER:
    for {
        msg, err := senzie.reader.ReadString(';')
        if err != nil {
            fmt.Println("Error reading: ", err.Error())

            break READER
        }

        // set read deadline to detect dead peers
	    senzie.conn.SetReadDeadline(time.Now().Add(readTimeout))

        println("received " + msg)

        // not handle TAK, TIK, TUK
        if(msg == "TAK;" || msg == "TIK;" || msg == "TUK;") {
            continue READER
        }

        // parse senz and handle it
        senz := parse(msg)
        if(senz.Receiver == config.switchName) {
            if(senz.Ztype == "GET") {
                // this is requesting pub key of other senzie
                // fing pubkey and send
                key := mongoStore.getKey(senz.Attr["name"])
                uid := senz.Attr["uid"]
                name := senz.Attr["name"]
                senzie.out <- keySenz(uid, key.Value, name, senzie.name)
            } else if(senz.Ztype == "AWA") {
                // this means message delivered to senzie
                // get senz with given uid
                uid := senz.Attr["uid"]
                var dz = mongoStore.dequeueSenzById(uid)

                // find sender and send GIYA
                if senzies[dz.Sender] != nil {
                    senzies[dz.Sender].out <- giyaSenz(uid, senzie.name)
                } else {
                    fmt.Println("no senzie to send giya: " + senz.Receiver)
                }
            }
        } else if(senz.Receiver == "*") {
            // broadcase senz
            // verify signature first of all
            payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
            senzieKey := getSenzieRsaPub(mongoStore.getKey(senzie.name).Value)
            err := verify(payload, senz.Digsig, senzieKey)
            if err != nil {
                println("cannot verify signarue, so dorp the conneciton")
                break READER
            }

            // send AWA back to sender
            uid := senz.Attr["uid"]
            senzie.out <- awaSenz(uid, senzie.name)

            // broadcast
            for k, v := range senzies {
                if (k != senz.Sender) {
                   v.out <- senz
                }
            }
        } else {
            // senz for another senzie
            // verify signature first of all
            payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
            senzieKey := getSenzieRsaPub(mongoStore.getKey(senzie.name).Value)
            err := verify(payload, senz.Digsig, senzieKey)
            if err != nil {
                println("cannot verify signarue, so dorp the conneciton")
                break READER
            }

            // send AWA back to sender
            uid := senz.Attr["uid"]
            senzie.out <- awaSenz(uid, senzie.name)

            // forwared senz msg to receiver
            if senzies[senz.Receiver] != nil {
                senzies[senz.Receiver].out <- senz
            } else {
                fmt.Println("no senzie to forward senz: ", senz.Receiver, " :"+ senz.Msg)
            }
        }
    }

    println("exit reader...")

    // quit all routeins of this senzie
    // close conn
    senzie.quit <- true
    senzie.conn.Close()
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

func writing(senzie *Senzie)  {
    // write
    WRITER:
    for {
        select {
        case <- senzie.quit:
            println("quiting/write/tick -- " + senzie.name)
            senzie.tik.Stop()
            break WRITER
        case senz := <-senzie.out:
            // enqueu senz, except AWA, GIYA and broadcase senz(receiver="*") 
            if (senz.Ztype != "AWA" && senz.Ztype != "GIYA" && senz.Receiver != "*") {
                mongoStore.enqueueSenz(senz)
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
