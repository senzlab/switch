package main
import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
)

type Senzie struct {
    name string
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
}

const (
    CONN_PORT = "7070"
    CONN_TYPE = "tcp"
)

var (
	senzies []*Senzie
    sz map[string]*Senzie
)

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
        senz, err := senzie.reader.ReadString(';')
        if err != nil {
            fmt.Println("Error reading: ", err.Error())
            return
        }

        // format senz
        var replacer = strings.NewReplacer(";", "", "\n", "")
        senz = strings.TrimSpace(replacer.Replace(senz))
        println(senz)

        if(senz == "SHARE") {
            println("SARE -- ")

            // senzie registered
            // todo set senzie name
            senzie.name = "eranga" 
            senzies = append(senzies, senzie)
            println(len(senzies))
        } else if(senz == "DATA") {
            println("DATA -- ")
            for _, senzie := range senzies {
                println("SENDING -- ")
                senzie.outgoing <- senz
            }
        }
    }
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
