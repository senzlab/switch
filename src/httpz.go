package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func promize(senz *Senz, from string, to string) {
	url := "http://" + chainzConfig.host + ":" + chainzConfig.port + "/promize"

	println("sending request " + url)

	body := []byte(senz.Msg)
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		println(err.Error())
		return
	}
	defer resp.Body.Close()

	// send

	println(resp.Status + "---")

	// handle response
	if resp.StatusCode == 201 {
		// means promize created
		msg, _ := ioutil.ReadAll(resp.Body)
		z := parse(string(msg))

		// send status to zfrom
		senzies[from].out <- statusSenz("SUCCESS", z.Attr["uid"], z.Attr["id"], from)

		// send response to zto
		senzies[to].out <- z
		println(string(msg))
	} else if resp.StatusCode == 200 {
		// premize redeemed
		// send response to zfrom
		msg, _ := ioutil.ReadAll(resp.Body)
		z := parse(string(msg))
		senzies[from].out <- z

		println(string(msg))
	} else {
		// means promize fail
		// send response to zfrom
		msg, _ := ioutil.ReadAll(resp.Body)
		z := parse(string(msg))
		senzies[from].out <- z

		println(string(msg))
	}
}
