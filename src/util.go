package main

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func parse(msg string) (*Senz, error) {
	fMsg := formatToParse(msg)
	tokens := strings.Split(fMsg, " ")
	senz := Senz{}
	senz.Msg = fMsg
	senz.Attr = map[string]string{}

	for i := 0; i < len(tokens); i++ {
		if i == 0 {
			senz.Ztype = tokens[i]
		} else if i == len(tokens)-1 {
			// signature at the end
			senz.Digsig = tokens[i]
		} else if strings.HasPrefix(tokens[i], "@") {
			// receiver @eranga
			senz.Receiver = tokens[i][1:]
		} else if strings.HasPrefix(tokens[i], "^") {
			// sender ^lakmal
			senz.Sender = tokens[i][1:]
		} else if strings.HasPrefix(tokens[i], "$") {
			// $key er2232
			key := tokens[i][1:]
			val := tokens[i+1]
			senz.Attr[key] = val
			i++
		} else if strings.HasPrefix(tokens[i], "#") {
			key := tokens[i][1:]
			nxt := tokens[i+1]

			if strings.HasPrefix(nxt, "#") || strings.HasPrefix(nxt, "$") ||
				strings.HasPrefix(nxt, "@") {
				// #lat #lon
				// #lat @eranga
				// #lat $key 32eewew
				senz.Attr[key] = ""
			} else {
				// #lat 3.2323 #lon 5.3434
				senz.Attr[key] = nxt
				i++
			}
		}
	}

	// set uid as the senz id
	senz.Uid = senz.Attr["uid"]

	// check for errors
	if senz.Sender == "" || senz.Receiver == "" || senz.Digsig == "" || senz.Ztype == "" || senz.Uid == "" || senz.Msg == "" {
		return nil, errors.New("Invalid senz")
	}

	return &senz, nil
}

func formatToParse(msg string) string {
	replacer := strings.NewReplacer(";", "", "\n", "", "\r", "")
	return strings.TrimSpace(replacer.Replace(msg))
}

func formatToSign(msg string) string {
	replacer := strings.NewReplacer(";", "", "\n", "", "\r", "", " ", "")
	return strings.TrimSpace(replacer.Replace(msg))
}

func uid() string {
	t := time.Now().UnixNano() / int64(time.Millisecond)
	return config.switchName + strconv.FormatInt(t, 10)
}

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func regSenz(uid string, status string, to string) Senz {
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

func keySenz(uid string, key string, name string, to string) Senz {
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

func awaSenz(uid string, to string) Senz {
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

func giyaSenz(uid string, to string) Senz {
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

func statusSenz(status string, uid string, to string) string {
	z := "DATA #status " + status +
		" #uid " + uid +
		" @" + to +
		" ^" + config.switchName
	s, _ := sign(z, getIdRsa())
	sz := z + " " + s
	return sz
}

func blobSenz(blob string, uid, to string) string {
	z := "DATA #blob " + blob +
		" #uid " + uid +
		" @" + to +
		" ^" + config.switchName
	s, _ := sign(z, getIdRsa())
	sz := z + " " + s
	return sz
}

func metaSenz(qSenz Senz, to string) string {
	z := "DATA #amount " + qSenz.Attr["amnt"] +
		" #uid " + qSenz.Attr["uid"] +
		" #id " + qSenz.Attr["id"] +
		" #from " + qSenz.Attr["from"] +
		" @" + to +
		" ^" + config.switchName
	s, _ := sign(z, getIdRsa())
	sz := z + " " + s
	return sz
}

func notifyPromizeSenz(senz *Senz) string {
	z := "DATA #uid " + senz.Attr["uid"] +
		" #id " + senz.Attr["id"] +
		" #amnt " + senz.Attr["amnt"] +
		" #from " + senz.Attr["from"] +
		" @" + senz.Receiver +
		" ^" + config.switchName
	//s, _ := sign(z, getIdRsa())
	s := "DIGSIG"
	sz := z + " " + s
	return sz
}

func notifyConnectSenz(senz *Senz) string {
	z := "DATA #uid " + senz.Attr["uid"] +
		" #pubkey " + senz.Attr["pubkey"] +
		" #from " + senz.Sender +
		" @" + senz.Attr["to"] +
		" ^" + config.switchName
	//s, _ := sign(z, getIdRsa())
	s := "DIGSIG"
	sz := z + " " + s
	return sz
}
