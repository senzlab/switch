package main
import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
)

type Senzie struct {
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
}

const (
    CONN_HOST = "localhost"
    CONN_PORT = "3333"
    CONN_TYPE = "tcp"
)

var (
	senzies []*Senzie
)

func main() {
    // listen for incoming conns
    l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }

    // close listern on app closes
    defer l.Close()
    fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

    for {
        // handle new connections 
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        go onConnect(conn) 
    }
}

func onConnect(conn net.Conn) {
    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    // read data from new conn
    for {
        senz, err := reader.ReadString(';')
        if err != nil {
            fmt.Println("Error reading: ", err.Error())
            conn.Close()
            return
        }

        // format senz
        var replacer = strings.NewReplacer(";", "", "\n", "")
        senz = strings.TrimSpace(replacer.Replace(senz))
        println(senz)

        if(senz == "SHARE") {
            println("SARE -- ")
            // client registered
            senzie := &Senzie {
                outgoing: make(chan string),
                reader: reader,
                writer: writer,
            }
            senzies = append(senzies, senzie)
            println(len(senzies))

            // start routing to write
            go writing(senzie)
        } else if(senz == "DATA") {
            println("DATA -- ")
            for _, senzie := range senzies {
                println("SENDING -- ")
                //senzie.writer.WriteString("hooo")
                //senzie.writer.Flush()
                senzie.outgoing <- senz
                //conn.Write([]byte(senz))
            }
        }
    }
}

func writing(senzie *Senzie)  {
    for senz := range senzie.outgoing {
        println("writing -- ")
        senzie.writer.WriteString(senz)
        senzie.writer.Flush()
    }
}
