package main

import (
    "strings"
    "strconv"
    "time"
)

func parse(msg string)Senz {
    fMsg := formatToParse(msg)
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

func formatToParse(msg string)string {
    replacer := strings.NewReplacer(";", "", "\n", "", "\r", "")
    return strings.TrimSpace(replacer.Replace(msg))
}

func formatToSign(msg string)string {
    replacer := strings.NewReplacer(";", "", "\n", "", "\r", "", " ", "")
    return strings.TrimSpace(replacer.Replace(msg))
}

func uid()string {
    t := time.Now().UnixNano() / int64(time.Millisecond)
    return config.switchName + strconv.FormatInt(t, 10)
}

func timestamp() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}

func regSenz(uid string, status string, to string)Senz {
    z := "DATA #status " + status +
                " #pubkey " + getIdRsaPubStr() +
                " #uid " + uid +
                " @" + to +
                " ^" + config.switchName
    s, _ := sign(z, getIdRsa())
    sm := z + " " + s

    sz := Senz{}
    sz.Ztype = "DATA"
    sz.Uid = uid
    sz.Msg = sm
    sz.Sender = config.switchName
    sz.Receiver = to

    return sz
}

func keySenz(uid string, key string, name string, to string)Senz {
    z := "DATA #pubkey " + key +
                " #name " + name +
                " #uid " + uid +
                " @" + to +
                " ^" + config.switchName
    s, _ := sign(z, getIdRsa())
    sm := z + " " + s

    sz := Senz{}
    sz.Ztype = "DATA"
    sz.Uid = uid
    sz.Msg = sm
    sz.Sender = config.switchName
    sz.Receiver = to

    return sz
}

func awaSenz(uid string, to string)Senz {
    z := "AWA #uid " + uid +
              " @" + to +
              " ^" + config.switchName
    s, _ := sign(z, getIdRsa())
    sm := z + " " + s

    sz := Senz{}
    sz.Ztype = "AWA"
    sz.Uid = uid
    sz.Msg = sm
    sz.Sender = config.switchName
    sz.Receiver = to

    return sz
}

func giyaSenz(uid string, to string)Senz {
    z := "GIYA #uid " + uid +
              " @" + to +
              " ^" + config.switchName
    s, _ := sign(z, getIdRsa())
    sm := z + " " + s

    sz := Senz{}
    sz.Ztype = "GIYA"
    sz.Uid = uid
    sz.Msg = sm
    sz.Sender = config.switchName
    sz.Receiver = to

    return sz
}

func scanSemiColon(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Return nothing if at end of file and no data passed
    if atEOF && len(data) == 0 {
        return 0, nil, nil
    }

    // Find the index of the input of ';'
    if i := strings.Index(string(data), ";"); i >= 0 {
        return i + 1, data[0:i], nil
    }

    // If at end of file with data return the data
    if atEOF {
        return len(data), data, nil
    }

    return
}
