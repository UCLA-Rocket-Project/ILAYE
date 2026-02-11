package commander

import (
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

type ShockAccelNum byte

const (
	SHOCK_ACCEL_1 ShockAccelNum = globals.CMD_GET_SHOCK_1_READING
	SHOCK_ACCEL_2 ShockAccelNum = globals.CMD_GET_SHOCK_2_READING
)

// tests
func CheckDigitalSDCommand(conn SerialReaderWriter, log io.Writer) bool {
	// enter inspect mode first
	fmt.Fprintf(log, "[Check Digital SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital SD]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Check Digital SD]: Dispatching sd card checker\n")
	firstUpdate := getSDUpdate(conn, log, globals.CMD_GET_DIGITAL_SD_UPDATE)

	if firstUpdate == nil {
		return false
	}

	fmt.Fprintf(log, "[Check Digital SD]: Entering normal mode\n")
	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital SD]: Failed to enter normal mode\n")
		return false
	}

	time.Sleep(SD_CARD_TEST_TIMEOUT)
	fmt.Fprintf(log, "[Check Digital SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital SD]: Failed to enter inspect mode\n")
		return false
	}

	time.Sleep(1 * time.Second)
	fmt.Fprintf(log, "[Check Digital SD]: Dispatching sd card checker again\n")
	secondUpdate := getSDUpdate(conn, log, globals.CMD_GET_DIGITAL_SD_UPDATE)

	if secondUpdate == nil {
		return false
	}

	return firstUpdate.FileSize < secondUpdate.FileSize && firstUpdate.LastTimestamp < secondUpdate.LastTimestamp
}

type shockData struct {
	AccX      float32
	AccY      float32
	AccZ      float32
	Timestamp uint32
}

func CheckDigitalShockCmd(conn SerialReaderWriter, log io.Writer, shockAccelNum ShockAccelNum) bool {
	var shockNum int

	switch shockAccelNum {
	case globals.CMD_GET_SHOCK_1_READING:
		shockNum = 1
	case globals.CMD_GET_SHOCK_2_READING:
		shockNum = 2
	}

	fmt.Fprintf(log, "[Check Digital %d]: Entering inspect mode\n", shockNum)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital %d]: Failed to enter inspect mode\n", shockNum)
		return false
	}

	shockUpdateMessage := getDispatchCommand(byte(shockAccelNum))
	conn.WriteSingleMessage(shockUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[Check Digital %d]: Sent command requesting Shock 1 update\n", shockNum)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Check Digital %d]: Read timed out", shockNum)
		return false
	}
	fmt.Fprintf(log, "[Check Digital %d]: Receieved response from boards of len %d\n", shockNum, len(res))
	streamReader := bytes.NewReader(res[:])
	var shockData shockData
	if err := binary.Read(streamReader, binary.LittleEndian, &shockData); err != nil {
		fmt.Fprintf(log, "[Check Digital %d]: Error decoding shock response\n", shockNum)
		return false
	}

	fmt.Fprintf(
		log,
		"[Check Digital %d]: Timestamp: %d Shock data: %f, %f, %f\n",
		shockNum, shockData.Timestamp, shockData.AccX, shockData.AccY, shockData.AccZ,
	)
	return true
}

// commands
func ClearDigitalSDCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Clear Digital SD]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Clear Digital SD]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Clear Digital SD]: sending command to clear digital SD card\n")

	cmd := getDispatchCommand(globals.CMD_CLEAR_ANALOG_SD)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Clear Digital SD]: Read timed out")
		return false
	}

	streamReader := bytes.NewReader(res[:])
	var freeSpace uint32
	if err := binary.Read(streamReader, binary.LittleEndian, &freeSpace); err != nil {
		fmt.Fprintf(log, "[Clear Digital SD]: Could not clear digital SD card\n")
		return false
	}

	fmt.Fprintf(log, "[Clear Digital SD]: Clear command acknowledged. Free space is now: %d MB\n", freeSpace)

	return true
}
