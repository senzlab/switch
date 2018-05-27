package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type AndroidNotification struct {
	Senz string `json:"senz"`
}

type FcmAndroid struct {
	To   string              `json:"to"`
	Data AndroidNotification `json:"data"`
}

type AppleNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Senz  string `json:"senz"`
}

type FcmApple struct {
	To               string            `json:"to"`
	ContentAvailable bool              `json:"content_available"`
	Notification     AppleNotification `json:"notification"`
}

func notifa(token string, n AndroidNotification) error {
	// marshel notification
	notification := FcmAndroid{
		To:   token,
		Data: n,
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
		println(err.Error())
		log.Printf("Error send fcm request: ", err.Error())
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

func notifi(token string, n AppleNotification) error {
	// marshel notification
	notification := FcmApple{
		To:               token,
		ContentAvailable: true,
		Notification:     n,
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
		println(err.Error())
		log.Printf("Error send fcm request: ", err.Error())
		return err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("fail notifi: ", resp.StatusCode, string(b))
		return errors.New("Invalid response")
	}

	log.Printf("success notifi response ", string(b))

	return nil
}
