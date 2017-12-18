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
	out         chan string
    quit        chan bool
    tik         *time.Ticker
    conn        net.Conn
}

type Senz struct {
    msg         string
    ztype       string
    sender      string
    receiver    string
    attr        map[string]string
    digsig      string
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

        // new senzie
        senzie := &Senzie {
            out: make(chan string),
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

        // not handle TAK, TIK, TUK
        if(msg == "TAK;" || msg == "TIK;" || msg == "TUK;") {
            continue READER
        }

        // parse senz and handle it
        senz := parse(msg)
        if(senz.receiver == config.switchName) {
            if(senz.ztype == "SHARE") {
                // this is shareing pub key(registration)
                // save pubkey in db
                senzie.name = senz.sender
                senzie.id = senz.attr["uid"]
                pubkey := senz.attr["pubkey"]
                key := mongoStore.get(senzie.name)

                println("SHARE pubKey to switch " + senzie.name + " " + senzie.id)

                if(key.Value == "") {
                    // not registerd senzie
                    // save pubkey
                    // add senzie
                    mongoStore.put(&Key{senzie.name, pubkey})
                    senzies[senzie.name] = senzie

                    // send status
                    z := "DATA #status REG_DONE #pubkey switchkey" +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    senzie.out <- z
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
                    z := "DATA #status REG_ALR #pubkey switchkey" +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    senzie.out <- z

                    // despatch queues messages of senzie
                    go despatching(senzie)
                } else {
                    // name already obtained
                    z := "DATA #status REG_FAIL #pubkey switchkey" +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    senzie.out <- z
                }
            } else if(senz.ztype == "GET") {
                // this is requesting pub key of other senzie
                // fing pubkey and send
                key := mongoStore.get(senz.attr["name"])
                z := "DATA #pubkey " + key.Value +
                            " #name " + senz.attr["name"] +
                            " #uid " + senz.attr["uid"] +
                            " @" + senzie.name +
                            " ^" + config.switchName +
                            " digisig"
                senzie.out <- z
            }
        } else {
            // senz for another senzie
            println("SENZ for senzie " + senz.msg)

            // send ack back to sender
            z := "DATA #status RECEIVED" +
                        " #uid " + senz.attr["uid"] +
                        " @" + senzie.name +
                        " ^" + config.switchName +
                        " digisig"
            senzie.out <- z

            // forwared senz msg to receiver
            if senzies[senz.receiver] != nil {
                senzies[senz.receiver].out <- senz.msg
            } else {
                println("no senzie " + senz.receiver)
            }
        }
    }
}

func despatching(senzie *Senzie) {
    // find queued messages from radis
    var zs = mongoStore.dequeueSenzByReceiver(senzie.name)

    // despatch queued messages to senzie
    for _, z := range zs {
        senzie.out <- z.msg
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
            println(senz)
            writer.WriteString(senz + ";")
            writer.Flush()
        case <-senzie.tik.C:
            println("ticking -- " + senzie.id)
            writer.WriteString("TIK;")
            writer.Flush()
        }
    }
}

func parse(msg string)*Senz {
    replacer := strings.NewReplacer(";", "", "\n", "")
    tokens := strings.Split(strings.TrimSpace(replacer.Replace(msg)), " ")
    senz := &Senz {}
    senz.msg = msg
    senz.attr = map[string]string{}

    for i := 0; i < len(tokens); i++ {
        if(i == 0) {
            senz.ztype = tokens[i]
        } else if(i == len(tokens) - 1) {
            // signature at the end
            senz.digsig = tokens[i]
        } else if(strings.HasPrefix(tokens[i], "@")) {
            // receiver @eranga
            senz.receiver = tokens[i][1:]
        } else if(strings.HasPrefix(tokens[i], "^")) {
            // sender ^lakmal
            senz.sender = tokens[i][1:]
        } else if(strings.HasPrefix(tokens[i], "$")) {
            // $key er2232
            key := tokens[i][1:]
            val := tokens[i + 1]
            senz.attr[key] = val
            i ++
        } else if(strings.HasPrefix(tokens[i], "#")) {
            key := tokens[i][1:]
            nxt := tokens[i + 1]

            if(strings.HasPrefix(nxt, "#") || strings.HasPrefix(nxt, "$") ||
                                                strings.HasPrefix(nxt, "@")) {
                // #lat #lon
                // #lat @eranga
                // #lat $key 32eewew
                senz.attr[key] = ""
            } else {
                // #lat 3.2323 #lon 5.3434
                senz.attr[key] = nxt
                i ++
            }
        }
    }

    return senz
}
