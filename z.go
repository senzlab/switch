package main
import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
)

type Senzie struct {
    name        string
	outgoing    chan string
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

const (
    CONN_PORT = "7070"
    CONN_TYPE = "tcp"
)

var senzies = map[string]*Senzie{}

func main() {
    // listen for incoming conns
    l, err := net.Listen(CONN_TYPE, ":" + CONN_PORT)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }

    // close listern on app closes
    defer l.Close()
    fmt.Println("Listening on " + CONN_PORT)

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
            return
        }

        // parse senz
        var senz = parse(senzMsg)

        if(senz.ztype == "SHARE") {
            println("SARE -- ")

            // senzie registered
            // todo set senzie name
            senzie.name = senz.sender
            senzies[senzie.name] = senzie
            println(len(senzies))
            println(senz.sender)
            println(senz.receiver)
            println(senzies[senzie.name].name)
        } else if(senz.ztype == "DATA") {
            println("DATA -- ")
            println(senz.sender)
            println(senz.receiver)

            // forwared senz
            var senzie = senzies[senz.receiver]
            senzie.outgoing <- senz.digsig
        }
    }

    // means senzie exists
}

func reading(senzie *Senzie) {
    // read senz
}

func writing(senzie *Senzie)  {
    for senz := range senzie.outgoing {
        println("writing -- ")
        senzie.writer.WriteString(senz)
        senzie.writer.Flush()
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
