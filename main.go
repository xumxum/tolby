package main

import (
	"log"
)

const verbose = true

func main() {
	log.SetFlags(log.Lmicroseconds)

	initConfiguration()
	initChatHistory()

	go runTelegramBot()

	// Infinite not blocking loop
	select {}

}
