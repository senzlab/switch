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
}

var config = Config {
    switchName: getEnv("SWITCH_NAME", "senzswitch"),
    switchPort: getEnv("SWITCH_PORT", "7070"),
    switchMode: getEnv("SWITCH_MODE", "dev"),
    mongoHost: getEnv("MONGO_HOST", "dev.localhost"),
    mongoPort: getEnv("MONGO_PORT", "27017"),
    mongoDb: getEnv("MONGO_DB", "senz"),
    keyColl: getEnv("KEY_COLL", "keys"),
    senzColl: getEnv("SENZ_COLL", "senzes"),
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }

    return fallback
}
