package main

import (
    "strings"
)

func formatToParse(msg string)string {
    replacer := strings.NewReplacer(";", "", "\n", "", "\r", "")
    return strings.TrimSpace(replacer.Replace(msg))
}

func formatToSign(msg string)string {
    replacer := strings.NewReplacer(";", "", "\n", "", "\r", "", " ", "")
    return strings.TrimSpace(replacer.Replace(msg))
}
