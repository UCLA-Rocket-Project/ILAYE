package main

import (
	"UCLA-Rocket-Project/ILAYE/internal/logger"
	"UCLA-Rocket-Project/ILAYE/internal/rpSerial"
	"UCLA-Rocket-Project/ILAYE/internal/terminal"
)

const LOG_FILE_PATH = "ILAYE.logs"
const BAUD_RATE = 115200

var STOP_SEQUENCE = []byte{'\r', '\n'}

func main() {
	log, err := logger.NewLogger(LOG_FILE_PATH)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	connector := func(port string) (terminal.SerialReaderWriter, error) {
		return rpSerial.NewRPSerial(port, BAUD_RATE, log), nil
	}

	terminal.StartApplication(rpSerial.ListPorts, connector, log)
}
