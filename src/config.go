package main

import (
	"os"
)

type Config struct {
	switchName string
	switchPort string
	switchMode string
	mongoHost  string
	mongoPort  string
	mongoDb    string
	keyColl    string
	senzColl   string
	dotKeys    string
	idRsa      string
	idRsaPub   string
	promizeApi string
	uzerApi    string
}

var config = Config{
	switchName: getEnv("ZWITCH_NAME", "senzswitch"),
	switchPort: getEnv("ZWITCH_PORT", "7171"),
	switchMode: getEnv("ZWITCH_MODE", "dev"),
	mongoHost:  getEnv("MONGO_HOST", "dev.localhost"),
	mongoPort:  getEnv("MONGO_PORT", "27017"),
	mongoDb:    getEnv("MONGO_DB", "senz"),
	keyColl:    getEnv("KEY_COLL", "keys"),
	senzColl:   getEnv("SENZ_COLL", "senzes"),
	dotKeys:    ".keys",
	idRsa:      ".keys/id_rsa",
	idRsaPub:   ".keys/id_rsa.pub",
	promizeApi: getEnv("PROMIZE_API", "https://chainz.com:8443/promizes"),
	uzerApi:    getEnv("UZER_API", "https://chainz.com:8443/uzers"),
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
