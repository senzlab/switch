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

type AndroidNotification struct {
	Senz string `json:"senz"`
}

type AppleNotification struct {
	Title string
	Type  string
	Senz  string
}

type Notification struct {
	To   string              `json:"to"`
	Data AndroidNotification `json:"data"`
}

func notifa(token string, an AndroidNotification) error {
	// marshel notification
	notification := Notification{
		To:   token,
		Data: an,
	}
	j, _ := json.Marshal(notification)
	log.Printf(string(j[:]))

	// request
	req, err := http.NewRequest("POST", fcmConfig.api, bytes.NewBuffer(j))
	if err != nil {
		log.Printf("Error init fcm request: ", err.Error)
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
		log.Printf("Error send fcm request: ", err.Error)
		return err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("fail notifa: ", resp.StatusCode, string(b))
		return errors.New("Invalid response")
	}

	log.Printf("success notifa response ", string(b))

	return nil
}

func notifi(client *apns2.Client, token string, an AppleNotification) {
	notification := &apns2.Notification{}
	notification.DeviceToken = token
	notification.Topic = apnConfig.topic
	payload := payload.NewPayload().Alert(an.Title).Badge(1).Custom(an.Type, an.Senz)
	notification.Payload = payload

	res, err := client.Push(notification)
	if err != nil {
		log.Printf("Error:", err)
	} else {
		log.Printf("%v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
	}
}
