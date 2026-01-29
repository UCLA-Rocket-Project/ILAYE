package main

import (
	"UCLA-Rocket-Project/ILAYE/internal/logger"
	"fmt"

	"go.bug.st/serial"
)

const LOG_FILE_PATH = "ILAYE.logs"

var STOP_SEQUENCE = []byte{'\r', '\n'}

func main() {
	log, err := logger.NewLogger(LOG_FILE_PATH)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ports, err := serial.GetPortsList()
	if err != nil {
		panic(err)
	}

	for _, port := range ports {
		fmt.Println(port)
	}
}
