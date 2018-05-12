package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func post(senz *Senz) ([]byte, int) {
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
		api = chainzConfig.promizeApi
	} else {
		api = chainzConfig.uzerApi
	}

	println("sending request " + api)

	req, err := http.NewRequest("POST", api, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")

	// send to senz api
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		println(err.Error())
		return nil, 400
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	println(string(b))

	return b, resp.StatusCode
}
