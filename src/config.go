package main

import (
	"os"
)

type Config struct {
	switchName string
	switchPort string
	switchMode string
	dotKeys    string
	idRsa      string
	idRsaPub   string
}

type MongoConfig struct {
	mongoHost string
	mongoPort string
	mongoDb   string
	username  string
	password  string
	keyColl   string
	senzColl  string
}

type ChainzConfig struct {
	name       string
	promizeApi string
	uzerApi    string
}

type FcmConfig struct {
	api       string
	serverKey string
}

type ApnConfig struct {
	api         string
	topic       string
	certificate string
}

var config = Config{
	switchName: getEnv("ZWITCH_NAME", "senzswitch"),
	switchPort: getEnv("ZWITCH_PORT", "7171"),
	switchMode: getEnv("ZWITCH_MODE", "dev"),
	dotKeys:    ".keys",
	idRsa:      ".keys/id_rsa",
	idRsaPub:   ".keys/id_rsa.pub",
}

var mongoConfig = MongoConfig{
	mongoHost: getEnv("MONGO_HOST", "dev.localhost"),
	mongoPort: getEnv("MONGO_PORT", "27017"),
	mongoDb:   "senz",
	username:  "senz",
	password:  "senz",
	keyColl:   "keys",
	senzColl:  "senzes",
}

var chainzConfig = ChainzConfig{
	name:       getEnv("CHAINZ_NAME", "sampath"),
	promizeApi: getEnv("PROMIZE_API", "https://chainz.com:8443/promizes"),
	uzerApi:    getEnv("UZER_API", "https://chainz.com:8443/uzers"),
}

var fcmConfig = FcmConfig{
	api:       getEnv("FCM_API", "https://fcm.googleapis.com/fcm/send"),
	serverKey: getEnv("FCM_SERVER_KEY", ""),
}

var apnConfig = ApnConfig{
	api:         getEnv("APN_API", "https://api.push.apple.com:443"),
	topic:       getEnv("APN_TOPIC", "com.creative.igift"),
	certificate: ".certs/apn.p12",
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
