package main

import (
	"fmt"
	"log"
	"os"
)

var logger *log.Logger

func reopenLogger() {
	if len(*logFilename) == 0 {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	} else {
		file, err := os.OpenFile(*logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0640))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't open log file: %s\n", err.Error())
			os.Exit(1)
		}
		logger = log.New(file, "", log.LstdFlags)
	}
}
