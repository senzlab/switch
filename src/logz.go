package main

import (
	"log"
	"os"
)

func initLogz() {
	// create logfile
	n := config.dotLogs + "/zwitch.log"
	f, err := os.OpenFile(n, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		os.Exit(1)
	}
	log.SetOutput(f)

	// log flags
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
