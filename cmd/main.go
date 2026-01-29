package main

import (
	"UCLA-Rocket-Project/ILAYE/internal/commander"
	"UCLA-Rocket-Project/ILAYE/internal/logger"
	"UCLA-Rocket-Project/ILAYE/internal/rpSerial"
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

	rpSerial := rpSerial.NewRPSerial("/dev/cu.usbserial-0001", 115200, log)

	for range 5 {
		fmt.Println(rpSerial.ReadSingleMessage())
		dispatchCommand := commander.GetDispatchCommand(commander.CMD_GET_ANALOG_SD_UPDATE)
		fmt.Printf("%p %s\n", &dispatchCommand, dispatchCommand)
		rpSerial.WriteSingleMessage(dispatchCommand[:], 4)
	}
}
