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
}

var config = Config {
    switchName: getEnv("SWITCH_NAME", "zswitch"),
    switchPort: getEnv("SWITCH_PORT", "7070"),
    switchMode: getEnv("SWITCH_MODE", "dev"),
    mongoHost: getEnv("MONGO_HOST", "dev.localhost"),
    mongoPort: getEnv("MONGO_PORT", "27017"),
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }

    return fallback
}
