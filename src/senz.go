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
    id          string
	out         chan Senz
    quit        chan bool
    tik         *time.Ticker
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

    for {
        // handle new connections
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        // enable keep alive
	    conn.(*net.TCPConn).SetKeepAlive(true)

        // new senzie
        senzie := &Senzie {
            out: make(chan Senz),
            quit: make(chan bool),
            tik: time.NewTicker(60 * time.Second),
            conn: conn,
        }

        go reading(senzie)
        go writing(senzie)
    }
}

func reading(senzie *Senzie) {
    reader := bufio.NewReader(senzie.conn)

    // read senz
    READER:
    for {
        msg, err := reader.ReadString(';')
        if err != nil {
            fmt.Println("Error reading: ", err.Error())

            // quit all routeins of this senzie
            senzie.quit <- true
            break READER
        }

        println("received " + msg + "from " + senzie.name)

        // not handle TAK, TIK, TUK
        if(msg == "TAK;" || msg == "TIK;" || msg == "TUK;") {
            continue READER
        }

        // parse senz and handle it
        senz := parse(msg)
        if(senz.Receiver == config.switchName) {
            if(senz.Ztype == "SHARE") {
                // this is shareing pub key(registration)
                // save pubkey in db
                senzie.name = senz.Sender
                senzie.id = senz.Attr["uid"]
                pubkey := senz.Attr["pubkey"]
                key := mongoStore.getKey(senzie.name)

                println("SHARE pubKey to switch " + senzie.name + " " + senzie.id)

                if(key.Value == "") {
                    // not registerd senzie
                    // save pubkey
                    // add senzie
                    mongoStore.putKey(&Key{senzie.name, pubkey})
                    senzies[senzie.name] = senzie

                    // send status
                    uid := senz.Attr["uid"]
                    senzie.out <- regSenz(uid, "REG_DONE", senzie.name)
                } else if(key.Value == pubkey) {
                    // already registerd senzie
                    // close existing senzie's conn
                    // delete existing senzie first
                    // then add new senzie
                    if senzies[senzie.name] != nil {
                        senzies[senzie.name].conn.Close()
                        delete(senzies, senzie.name)
                    }
                    senzies[senzie.name] = senzie

                    // send status
                    uid := senz.Attr["uid"]
                    senzie.out <- regSenz(uid, "REG_ALR", senzie.name) 

                    // dispatch queued messages of senzie
                    go dispatching(senzie)
                } else {
                    // name already obtained
                    // send status
                    uid := senz.Attr["uid"]
                    senzie.out <- regSenz(uid, "REG_FAIL", senzie.name)
                }
            } else if(senz.Ztype == "GET") {
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
        } else {
            // senz for another senzie
            println("SENZ for senzie " + senz.Msg)

            // verify signature first of all
            payload := strings.Replace(senz.Msg, senz.Digsig, "", -1)
            senzieKey := getSenzieRsaPub(mongoStore.getKey(senzie.name).Value)
            err := verify(payload, senz.Digsig, senzieKey)
            if err != nil {
                println("cannot verify signarue, so dorp the conneciton")
                senzie.quit <- true
                break READER
            }

            // send AWA back to sender
            uid := senz.Attr["uid"]
            senzie.out <- awaSenz(uid, senzie.name)

            // we queue the senz
            mongoStore.enqueueSenz(senz)

            // forwared senz msg to receiver
            if senzies[senz.Receiver] != nil {
                senzies[senz.Receiver].out <- senz
            } else {
                fmt.Println("no senzie to forward senz: ", senz.Receiver, " :"+ senz.Msg)
            }
        }
    }
}

func dispatching(senzie *Senzie) {
    // find queued messages from mongo store
    var zs = mongoStore.dequeueSenzByReceiver(senzie.name)

    // dispatch queued messages to senzie
    for _, z := range zs {
        senzie.out <- z
    }
}

func writing(senzie *Senzie)  {
    writer := bufio.NewWriter(senzie.conn)

    // write
    WRITER:
    for {
        select {
        case <- senzie.quit:
            println("quiting/write/tick -- " + senzie.id)
            senzie.tik.Stop()
            break WRITER
        case senz := <-senzie.out:
            println("writing -- " + senzie.id)
            println(senz.Msg)
            writer.WriteString(senz.Msg + ";")
            writer.Flush()
        case <-senzie.tik.C:
            println("ticking -- " + senzie.id)
            writer.WriteString("TIK;")
            writer.Flush()
        }
    }
}
