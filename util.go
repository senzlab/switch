package main

import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
)

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
