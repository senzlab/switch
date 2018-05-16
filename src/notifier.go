package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"io/ioutil"
	"log"
	"net/http"
)

type Notification struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

func notifa(to string, msg string) error {
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

func notifi(client *apns2.Client, token string, key string, z string) {
	notification := &apns2.Notification{}
	notification.DeviceToken = token
	notification.Topic = "com.creative.igift"
	payload := payload.NewPayload().Alert("New iGift").Badge(1).Custom("senz_connect", z)
	notification.Payload = payload

	res, err := client.Push(notification)
	if err != nil {
		log.Fatal("Error:", err)
	}

	log.Printf("%v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
}
