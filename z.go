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
	outgoing    chan string
    pinging     chan string
    quit        chan bool
	reader      *bufio.Reader
	writer      *bufio.Writer
}

type Senz struct {
    ztype       string
    sender      string
    receiver    string
    attr        map[string]string
    digsig   string
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
    keyStore.put(Key{name: "eranga", value: "we2323"})
    println(keyStore.get("eranga").name)

    for {
        // handle new connections 
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        // new senzie
        senzie := &Senzie {
            outgoing: make(chan string),
            pinging: make(chan string),
            quit: make(chan bool),
            reader: bufio.NewReader(conn),
            writer: bufio.NewWriter(conn),
        }
        go listening(senzie)
        go writing(senzie)
    }
}

func listening(senzie *Senzie)  {
    // read data
    for {
        senzMsg, err := senzie.reader.ReadString(';')
        if err != nil {
            fmt.Println("Error reading: ", err.Error())
            break
        }

        // parse senz
        var senz = parse(senzMsg)

        if(senz.ztype == "SHARE") {
            println("SHARE -- ")

            // senzie registered
            senzie.name = senz.sender
            senzies[senzie.name] = senzie

            // start pinging
            go pinging(senzie)
        } else if(senz.ztype == "DATA") {
            println("DATA -- ")

            // forwared senz
            var senzie = senzies[senz.receiver]
            senzie.outgoing <- senz.digsig
        }
    }

    // senzie exists
    // quit all routeins
    senzie.quit <- true
}

func reading(senzie *Senzie) {
    // read senz
}

func pinging(senzie *Senzie) {
    // ping
    for {
        select {
        case <- senzie.quit:
            println("quiting -- ")
            break
        default:
            <-time.After(120 * time.Second)
            senzie.pinging <- "TIK"
        }
    }
}

func writing(senzie *Senzie)  {
    // write
    for {
        select {
        case <- senzie.quit:
            println("quiting -- ")
            break
        case senz := <-senzie.outgoing:
            println("writing -- ")
            senzie.writer.WriteString(senz)
            senzie.writer.Flush()
        case <- senzie.pinging:
            println("pinging -- ")
        }
    }
}

func parse(senzMsg string)*Senz {
    var replacer = strings.NewReplacer(";", "", "\n", "")
    var tokens = strings.Split(strings.TrimSpace(replacer.Replace(senzMsg)), " ")
    var senz = &Senz {}

    for i := 0; i < len(tokens); i++ {
        if(i == 0) {
            senz.ztype = tokens[i]
        } else if(i == len(tokens) - 1) {
            senz.digsig = tokens[i]
        } else if(strings.HasPrefix(tokens[i], "@")) {
            senz.receiver = tokens[i][1:]
        } else if(strings.HasPrefix(tokens[i], "^")) {
            senz.sender = tokens[i][1:]
        }
    }

    return senz
}
