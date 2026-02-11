package commander

import (
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// tests
func CheckAnalogSDCommand(conn SerialReaderWriter, log io.Writer) bool {
	// enter inspect mode first
	fmt.Fprintf(log, "[Check Analog SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Analog SD]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Check Analog SD]: Dispatching sd card checker\n")
	firstUpdate := getSDUpdate(conn, log, globals.CMD_GET_ANALOG_SD_UPDATE)

	if firstUpdate == nil {
		return false
	}

	fmt.Fprintf(log, "[Check Analog SD]: Entering normal mode\n")
	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Check Analog SD]: Failed to enter normal mode\n")
		return false
	}

	time.Sleep(SD_CARD_TEST_TIMEOUT)
	fmt.Fprintf(log, "[Check Analog SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Analog SD]: Failed to enter inspect mode\n")
		return false
	}

	time.Sleep(1 * time.Second)
	fmt.Fprintf(log, "[Check Analog SD]: Dispatching sd card checker again\n")
	secondUpdate := getSDUpdate(conn, log, globals.CMD_GET_ANALOG_SD_UPDATE)

	if secondUpdate == nil {
		return false
	}

	return firstUpdate.FileSize < secondUpdate.FileSize && firstUpdate.LastTimestamp < secondUpdate.LastTimestamp
}

func CheckAnalogLCCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Check Analog LC]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Analog LC]: Failed to enter inspect mode\n")
		return false
	}

	sdUpdateMessage := getDispatchCommand(globals.CMD_GET_ANALOG_LC_READING)
	conn.WriteSingleMessage(sdUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[Chceck Analog LC]: Sent command requesting LC update\n")

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Check Analog LC]: Read timed out")
		return false
	}
	fmt.Fprintf(log, "[Check Analog LC]: Receieved response from boards\n")
	streamReader := bytes.NewReader(res[:])
	var updateData float32
	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
		return false
	}
	fmt.Fprintf(log, "[Check Analog LC]: Raw Reading: %f, Calibrated Reading: %f\n", updateData, -223810.211*updateData+10.86155)

	return true
}

// commands
func ClearAnalogSDCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Clear Analog SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Clear Analog SD]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Clear Analog SD]: sending command to clear analog SD card\n")

	cmd := getDispatchCommand(globals.CMD_CLEAR_ANALOG_SD)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Clear Analog SD]: Read timed out")
		return false
	}

	streamReader := bytes.NewReader(res[:])
	var freeSpace uint32
	if err := binary.Read(streamReader, binary.LittleEndian, &freeSpace); err != nil {
		fmt.Fprintf(log, "[Clear Analog SD]: Could not clear analog SD card\n")
		return false
	}

	fmt.Fprintf(log, "[Clear Analog SD]: Clear command acknowledged. Free space is now: %d MB\n", freeSpace)

	return true
}
