package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type SenzMsg struct {
	Uid string
	Msg string
}

func promize(senz *Senz, from string, to string) {
	url := "http://" + chainzConfig.host + ":" + chainzConfig.port + "/promize"

	println("sending request " + url)
	// marshel senz
	senzMsg := SenzMsg{
		Uid: senz.Attr["uid"],
		Msg: senz.Msg,
	}
	j, _ := json.Marshal(senzMsg)

	// send to senz api
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(j))
	if err != nil {
		println(err.Error())
		return
	}
	defer resp.Body.Close()

	println(resp.Status + "---")

	// handle response
	if resp.StatusCode == 201 {
		// means promize created
		b, _ := ioutil.ReadAll(resp.Body)

		println(string(b))

		// unmarshel senz response
		var zmsgs []SenzMsg
		json.Unmarshal(b, &zmsgs)

		// iterate over each and every msg and process it
		for _, zmsg := range zmsgs {
			z := parse(string(zmsg.Msg))
			// TODO check senzie exists
			senzies[z.Receiver].out <- z
		}
	} else if resp.StatusCode == 200 {
		// premize redeemed
		// send response to zfrom
		b, _ := ioutil.ReadAll(resp.Body)

		println(string(b))

		// unmarshel senz response
		var zmsgs []SenzMsg
		json.Unmarshal(b, &zmsgs)

		// iterate over each and every msg and process it
		for _, zmsg := range zmsgs {
			z := parse(string(zmsg.Msg))
			// TODO check senzie exists
			senzies[z.Receiver].out <- z
		}
	} else {
		// means promize fail
		// send response to zfrom
		msg, _ := ioutil.ReadAll(resp.Body)
		z := parse(string(msg))
		senzies[from].out <- z

		println(string(msg))
	}
}
