package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type Notification struct {
	To   string  `json:"to"`
	Data SenzMsg `json:"data"`
}

func notify(to string, msg SenzMsg) error {
	// marshel notification
	notification := Notification{
		To:   to,
		Data: msg,
	}
	j, _ := json.Marshal(notification)
	println(string(j[:]))

	// request
	req, err := http.NewRequest("POST", fcmConfig.api, bytes.NewBuffer(j))
	if err != nil {
		println(err.Error)
		return err
	}

	// headers
	key := "key=" + fcmConfig.serverKey
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", key)

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		println(err.Error)
		return err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	println(resp.StatusCode)
	println(string(b))

	if resp.StatusCode != 200 {
		println("invalid response")
		return errors.New("Invalid response")
	}

	return nil
}
