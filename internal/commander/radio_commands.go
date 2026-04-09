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
func CheckRadioSDCommand(conn SerialReaderWriter, log io.Writer) bool {
	// enter inspect mode first
	fmt.Fprintf(log, "[Check Radio SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Radio SD]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Check Radio SD]: Dispatching sd card checker\n")
	firstUpdate := getSDUpdate(conn, log, globals.CMD_GET_RADIO_SD_UPDATE)

	if firstUpdate == nil {
		return false
	}

	fmt.Fprintf(log, "[Check Radio SD]: Entering normal mode\n")
	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Check Radio SD]: Failed to enter normal mode\n")
		return false
	}

	time.Sleep(SD_CARD_TEST_TIMEOUT)
	fmt.Fprintf(log, "[Check Radio SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Radio SD]: Failed to enter inspect mode\n")
		return false
	}

	time.Sleep(1 * time.Second)
	fmt.Fprintf(log, "[Check Radio SD]: Dispatching sd card checker again\n")
	secondUpdate := getSDUpdate(conn, log, globals.CMD_GET_RADIO_SD_UPDATE)

	if secondUpdate == nil {
		return false
	}

	return firstUpdate.FileSize <= secondUpdate.FileSize && firstUpdate.LastTimestamp <= secondUpdate.LastTimestamp
}

// Commands
func ClearRadioSDCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Clear Radio SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Clear Radio SD]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Clear Radio SD]: sending command to clear Radio SD card\n")

	cmd := getDispatchCommand(globals.CMD_CLEAR_RADIO_SD)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Clear Radio SD]: Read timed out")
		return false
	}

	streamReader := bytes.NewReader(res[:])
	var freeSpace uint32
	if err := binary.Read(streamReader, binary.LittleEndian, &freeSpace); err != nil {
		fmt.Fprintf(log, "[Clear Radio SD]: Could not clear Radio SD card\n")
		return false
	}

	fmt.Fprintf(log, "[Clear Radio SD]: Clear command acknowledged. Free space is now: %d MB\n", freeSpace)

	return true
}

// radios have a different time format
type radioSdUpdate struct {
	LastTimestamp uint32
	FileSize      uint32
}

func getRadioSDUpdate(conn SerialReaderWriter, log io.Writer, command byte) *radioSdUpdate {
	sdUpdateMessage := getDispatchCommand(command)
	conn.WriteSingleMessage(sdUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[SD Update]: Sent command requesting SD card update\n")

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[SD Update]: Read timed out")
		return nil
	}
	fmt.Fprintf(log, "[SD Update]: Receieved response from boards\n")
	streamReader := bytes.NewReader(res[:])
	var updateData radioSdUpdate
	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
		fmt.Fprintf(log, "[SD Update]: Could not decode board response, %s\n", err)
		return nil
	}
	fmt.Fprintf(log, "[SD Update]: file size: %d, last update timestamp: %d\n", updateData.FileSize, updateData.LastTimestamp)

	return &updateData
}

func InspectRadioSDCard(conn SerialReaderWriter, log io.Writer, boardType string, command byte) bool {
	fmt.Fprintf(log, "[Check %s SD]: Entering inspect mode\n", boardType)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check %s SD]: Failed to enter inspect mode\n", boardType)
		return false
	}

	fmt.Fprintf(log, "[Check %s SD]: Dispatching sd card checker\n", boardType)
	firstUpdate := getRadioSDUpdate(conn, log, command)

	if firstUpdate == nil {
		return false
	}

	fmt.Fprintf(log, "[Check %s SD]: Entering normal mode\n", boardType)
	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Check %s SD]: Failed to enter normal mode\n", boardType)
		return false
	}

	time.Sleep(SD_CARD_TEST_TIMEOUT)
	fmt.Fprintf(log, "[Check %s SD]: Entering inspect mode\n", boardType)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check %s SD]: Failed to enter inspect mode\n", boardType)
		return false
	}

	time.Sleep(1 * time.Second)
	fmt.Fprintf(log, "[Check %s SD]: Dispatching sd card checker again\n", boardType)
	secondUpdate := getRadioSDUpdate(conn, log, command)

	if secondUpdate == nil {
		return false
	}

	return true
}
