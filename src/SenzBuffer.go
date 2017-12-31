package main

import (
    "sync"
    "bytes"
    "net"
    "os"
)

type SenzBuffer struct {
    b bytes.Buffer
    m sync.Mutex
}

func (b *SenzBuffer) Read(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Read(p)
}

func (b *SenzBuffer) Write(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Write(p)
}

func (b *SenzBuffer) String() string {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.String()
}

func (b *SenzBuffer) writing(conn net.Conn) {
    tmp := make([]byte, 256)

    for {
        n, err := conn.Read(tmp)
        if err != nil {
            println(err.Error())
            os.Exit(1)
        }

        if n > 0 {
            b.b.Write(tmp[:n])
        }
    }
}

func (b *SenzBuffer) reading(senzie *Senzie) {
    // reader
    READER:
    for {
        select {
        case <- senzie.quit:
            break READER
        default:
            l, _ := b.b.ReadBytes(';')
            if(len(l) > 0) {
                senz := string(l[:])
                println(senz)
            }
        }
    }
}
