package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type SenzMsg struct {
	Uid string
	Msg string
}

func promize(senz *Senz) {

	// load client cert
	cert, err := tls.LoadX509KeyPair(".certs/client.crt", ".certs/client.key")
	if err != nil {
		println(err.Error())
	}

	// load CA cert
	caCert, err := ioutil.ReadFile(".certs/ca.crt")
	if err != nil {
		println(err.Error())
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// https client tls config
	// InsecureSkipVerify true means not validate server certificate (so no need to set RootCAs)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		//RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	// marshel senz
	senzMsg := SenzMsg{
		Uid: senz.Attr["uid"],
		Msg: senz.Msg,
	}
	j, _ := json.Marshal(senzMsg)

	// post request
	var api string
	if _, ok := senz.Attr["blob"]; ok {
		api = config.promizeApi
	} else {
		api = config.uzerApi
	}

	println("sending request " + api)

	req, err := http.NewRequest("POST", api, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")

	// send to senz api
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		println(err.Error())
		return
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	println(string(b))

	// unmarshel senz response
	var zmsgs []SenzMsg
	json.Unmarshal(b, &zmsgs)

	// iterate over each and every msg and process it
	for _, zmsg := range zmsgs {
		z := parse(string(zmsg.Msg))

		// check senzie exists
		if senzies[z.Receiver] != nil {
			senzies[z.Receiver].out <- z
		} else {
			println("no senzie to send httpz senz, enqueued " + z.Msg)
			mongoStore.enqueueSenz(&z)
		}
	}
}
