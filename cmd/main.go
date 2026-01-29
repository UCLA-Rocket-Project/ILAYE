package main

import (
	"UCLA-Rocket-Project/ILAYE/internal/logger"
	"UCLA-Rocket-Project/ILAYE/internal/rpSerial"
	"UCLA-Rocket-Project/ILAYE/internal/terminal"
)

const LOG_FILE_PATH = "ILAYE.logs"

var STOP_SEQUENCE = []byte{'\r', '\n'}

func main() {
	log, err := logger.NewLogger(LOG_FILE_PATH)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// Connector adapter
	connector := func(port string) (terminal.SerialConnection, error) {
		return rpSerial.NewRPSerial(port, 115200, log), nil
	}

	terminal.Start(rpSerial.GetPortsList, connector)
}
