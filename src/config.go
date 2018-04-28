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
	keyColl   string
	senzColl  string
}

type ChainzConfig struct {
	name       string
	key        string
	promizeApi string
	uzerApi    string
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
	keyColl:   "keys",
	senzColl:  "senzes",
}

var chainzConfig = ChainzConfig{
	name:       getEnv("CHAINZ_NAME", "sampath"),
	key:        getEnv("CHAINZ_KEY", "chainzkey"),
	promizeApi: getEnv("PROMIZE_API", "https://chainz.com:8443/promizes"),
	uzerApi:    getEnv("UZER_API", "https://chainz.com:8443/uzers"),
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
