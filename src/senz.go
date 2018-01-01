package main

import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
    "strconv"
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
    Uid          string
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
                    uid := uid()
                    z := "DATA #status REG_DONE #pubkey switchkey" +
                                " #uid " + uid +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    sz := Senz{}
                    sz.Uid = uid 
                    sz.Msg = z
                    sz.Sender = config.switchName
                    sz.Receiver = senzie.name

                    senzie.out <- sz 
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
                    uid := uid()
                    z := "DATA #status REG_ALR #pubkey switchkey" +
                                " #uid " + uid +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    sz := Senz{}
                    sz.Uid = uid
                    sz.Msg = z
                    sz.Sender = config.switchName
                    sz.Receiver = senzie.name

                    senzie.out <- sz

                    // dispatch queued messages of senzie
                    go dispatching(senzie)
                } else {
                    // name already obtained
                    uid := uid()
                    z := "DATA #status REG_FAIL #pubkey switchkey" +
                                " #uid " + uid +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    sz := Senz{}
                    sz.Uid = uid
                    sz.Msg = z
                    sz.Sender = config.switchName
                    sz.Receiver = senzie.name

                    senzie.out <- sz
                }
            } else if(senz.Ztype == "GET") {
                // this is requesting pub key of other senzie
                // fing pubkey and send
                key := mongoStore.getKey(senz.Attr["name"])
                uid := senz.Attr["uid"]
                z := "DATA #pubkey " + key.Value +
                            " #name " + senz.Attr["name"] +
                            " #uid " + senz.Attr["uid"] +
                            " @" + senzie.name +
                            " ^" + config.switchName +
                            " digisig"
                sz := Senz{}
                sz.Uid = uid
                sz.Msg = z
                sz.Sender = config.switchName
                sz.Receiver = senzie.name

                senzie.out <- sz
            } else if(senz.Ztype == "AWA") {
                // this means message delivered to senzie
                // get senz with given uid
                uid := senz.Attr["uid"]
                var dz = mongoStore.dequeueSenzById(uid)

                // giya message 
                z := "GIYA #uid " + uid +
                            " @" + senzie.name +
                            " ^" + config.switchName +
                            " digisig"
                sz := Senz{}
                sz.Uid = uid
                sz.Msg = z
                sz.Sender = config.switchName
                sz.Receiver = dz.Sender

                // find sender and send GIYA 
                if senzies[dz.Sender] != nil {
                    senzies[dz.Sender].out <- sz
                } else {
                    fmt.Println("no senzie to send giya: " + senz.Receiver, " :" + sz.Msg)
                }
            }
        } else {
            // senz for another senzie
            println("SENZ for senzie " + senz.Msg)

            // send AWA back to sender
            uid := senz.Attr["uid"]
            z := "AWA #uid " + uid +
                        " @" + senzie.name +
                        " ^" + config.switchName +
                        " digisig"
            sz := Senz{}
            sz.Uid = uid
            sz.Msg = z
            sz.Sender = config.switchName
            sz.Receiver = senzie.name

            senzie.out <- sz

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

func parse(msg string)Senz {
    replacer := strings.NewReplacer(";", "", "\n", "", "\r", "")
    fMsg := strings.TrimSpace(replacer.Replace(msg))
    tokens := strings.Split(fMsg, " ")
    senz := Senz {}
    senz.Msg = fMsg
    senz.Attr = map[string]string{}

    for i := 0; i < len(tokens); i++ {
        if(i == 0) {
            senz.Ztype = tokens[i]
        } else if(i == len(tokens) - 1) {
            // signature at the end
            senz.Digsig = tokens[i]
        } else if(strings.HasPrefix(tokens[i], "@")) {
            // receiver @eranga
            senz.Receiver = tokens[i][1:]
        } else if(strings.HasPrefix(tokens[i], "^")) {
            // sender ^lakmal
            senz.Sender = tokens[i][1:]
        } else if(strings.HasPrefix(tokens[i], "$")) {
            // $key er2232
            key := tokens[i][1:]
            val := tokens[i + 1]
            senz.Attr[key] = val
            i ++
        } else if(strings.HasPrefix(tokens[i], "#")) {
            key := tokens[i][1:]
            nxt := tokens[i + 1]

            if(strings.HasPrefix(nxt, "#") || strings.HasPrefix(nxt, "$") ||
                                                strings.HasPrefix(nxt, "@")) {
                // #lat #lon
                // #lat @eranga
                // #lat $key 32eewew
                senz.Attr[key] = ""
            } else {
                // #lat 3.2323 #lon 5.3434
                senz.Attr[key] = nxt
                i ++
            }
        }
    }

    // set uid as the senz id
    senz.Uid = senz.Attr["uid"]

    return senz
}

func uid()string {
    t := time.Now().UnixNano() / int64(time.Millisecond)
    return config.switchName + strconv.FormatInt(t, 10)
}
