package main

import "log"

func DBG(msg string) {
	if *cnfVerbose {
		log.Println("DEBUG: " + msg)
	}
}

func INF(msg string) {
	log.Println("INFO: " + msg)
}

func WRN(msg string) {
	log.Println("WARNING: " + msg)
}
