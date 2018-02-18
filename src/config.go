package main

import (
    "os"
)

type Config struct {
	switchName  string
    switchPort  string
    switchMode  string
    mongoHost   string
    mongoPort   string
    mongoDb     string
    keyColl     string
    senzColl    string
    dotKeys     string
    idRsa       string
    idRsaPub    string
}

var config = Config {
    switchName: getEnv("ZWITCH_NAME", "senzswitch"),
    switchPort: getEnv("ZWITCH_PORT", "7070"),
    switchMode: getEnv("ZWITCH_MODE", "dev"),
    mongoHost: getEnv("MONGO_HOST", "dev.localhost"),
    mongoPort: getEnv("MONGO_PORT", "27017"),
    mongoDb: getEnv("MONGO_DB", "senz"),
    keyColl: getEnv("KEY_COLL", "keys"),
    senzColl: getEnv("SENZ_COLL", "senzes"),
    dotKeys: getEnv("DOT_KEYS", ".keys"),
    idRsa: getEnv("ID_RSA", ".keys/id_rsa"),
    idRsaPub: getEnv("ID_RSA_PUB", ".keys/id_rsa.pub"),
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }

    return fallback
}
