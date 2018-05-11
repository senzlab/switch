package main

import "fmt"

func mmain() {
	keyStr := "MIGJAoGBAO9l/TgQJlZzAOGmrwgnOik95657fEEanWyv+T7w2zAhaV8W7P/r39V/5d/Igm2owpgXGZZN+HxA3HOcXuMho/AubT5CiiKfuXxg68dlpj9rKe8FjEOQa26/l8YbZx1XPMhEsd/IMntMOZBlyDlvEVsI3o4u4NgNKvvfXSbmJD3RAgMBAAE="
	key := getSenzieRsaPub(keyStr)
	if key != nil {
		println("hooooo")
	}

	sig := "gI1g3e8GluSQVxHv8ztEgAHYA2DIDovXvtMobW9vzOPpI+ZeTGMnank3/YZG9exI9Zr5jo1UH+UyuVOfML00a3IoMd1n+Nr6mGVlHaNRXE0bBnG0LIZrR6eV3b58JMb/Q/3/CsAl31puvv+JD43FPOq/Hfe5IrpC0ZMdL7GTAK8="
	msg := "SHARE #uid eranga1526004960021 #pubkey MIGJAoGBAO9l/TgQJlZzAOGmrwgnOik95657fEEanWyv+T7w2zAhaV8W7P/r39V/5d/Igm2owpgXGZZN+HxA3HOcXuMho/AubT5CiiKfuXxg68dlpj9rKe8FjEOQa26/l8YbZx1XPMhEsd/IMntMOZBlyDlvEVsI3o4u4NgNKvvfXSbmJD3RAgMBAAE= @senzswitch ^eranga"

	err := verify(msg, sig, key)
	if err != nil {
		fmt.Println("Error sign:", err.Error())
	}
}
