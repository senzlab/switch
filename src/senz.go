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
var keyStore = &KeyStore{}

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
    keyStore.session = session

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

            // senzie exists
            // quit all routeins
            senzie.quit <- true

            break READER
        }

        // not handle TAK, TIK, TUK
        if(msg == "TAK;" || msg == "TIK;" || msg == "TUK;") {
            continue READER
        }

        // parse senz and handle it
        var senz = parse(msg)
        if(senz.receiver == config.switchName) {
            if(senz.ztype == "SHARE") {
                // this is shareing pub key(registration)
                println("SHARE pubKey to switch")

                // save pubkey in db
                senzie.name = senz.sender
                pubkey := senz.attr["pubkey"]
                key := keyStore.get(senzie.name)

                if(key.Value == "") {
                    // not registerd senzie
                    // save pubkey
                    // send status
                    keyStore.put(&Key{senzie.name, pubkey})
                    z := "DATA #status REG_DONE #pubkey switchkey" +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    senzie.out <- z

                    // senzie registered
                    // take existing senzie and stop it
                    // add new senzie
                    if rSenzie, ok := senzies[senzie.name]; ok {
                        rSenzie.conn.Close()
                    }
                    senzies[senzie.name] = senzie
                } else if(key.Value == pubkey) {
                    // re sharing pubkey
                    // send status
                    z := "DATA #status REG_ALR #pubkey switchkey" +
                                " @" + senzie.name +
                                " ^" + config.switchName +
                                " digisig"
                    senzie.out <- z

                    // senzie registered
                    // take existing senzie and stop it
                    // add new senzie
                    if rSenzie, ok := senzies[senzie.name]; ok {
                        rSenzie.conn.Close()
                    }
                    senzies[senzie.name] = senzie
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
                key := keyStore.get(senz.attr["name"])
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
            senzies[senz.receiver].out <- senz.msg
        }
    }
}

func writing(senzie *Senzie)  {
    writer := bufio.NewWriter(senzie.conn)

    // write
    WRITER:
    for {
        select {
        case <- senzie.quit:
            println("quiting/write/tick -- ")
            senzie.tik.Stop()
            break WRITER
        case senz := <-senzie.out:
            println("writing -- ")
            println(senz)
            writer.WriteString(senz + ";")
            writer.Flush()
        case <-senzie.tik.C:
            println("ticking -- ")
            writer.WriteString("TIK;")
            writer.Flush()
        }
    }
}

func parse(msg string)*Senz {
    var replacer = strings.NewReplacer(";", "", "\n", "")
    var tokens = strings.Split(strings.TrimSpace(replacer.Replace(msg)), " ")
    var senz = &Senz {}
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
