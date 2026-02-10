package commander

import (
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const COMMAND_SEQUENCE_SIZE = 3
const COMMAND_BYTE_IDX = 2

const SD_CARD_TEST_TIMEOUT = 5 * time.Second

type SerialReaderWriter interface {
	WriteSingleMessage(message []byte, size int)
	ReadSingleOrTimeout() ([]byte, error)
}

func getDispatchCommand(cmd byte) [COMMAND_SEQUENCE_SIZE]byte {
	// for consistency with terminal, use carraige return when sending back a command
	return [COMMAND_SEQUENCE_SIZE]byte{cmd, '\r', '\n'}
}

// need some sort of verification for the commands
func EnterNormalCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Enter Normal Command]: sending command to enter normal mode\n")

	cmd := getDispatchCommand(globals.CMD_ENTER_NORMAL)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Enter Normal Command]: Read timed out")
		return false
	}

	if res[0] == globals.CMD_ENTER_NORMAL {
		fmt.Fprintf(log, "[Enter Normal Command]: Normal mode transition acknowledged\n")
	} else {
		fmt.Fprintf(log, "[Enter Normal Command]: Could not enter normal mode")
	}

	return res[0] == globals.CMD_ENTER_NORMAL
}

// need some more verification for this
func EnterInspectCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Enter Inspect Command]: sending command to enter inspect mode\n")

	cmd := getDispatchCommand(globals.CMD_ENTER_INSPECT)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Enter Inspect Command]: Read timed out")
		return false
	}

	if res[0] == globals.CMD_ENTER_INSPECT {
		fmt.Fprintf(log, "[Enter Inspect Command]: Inspect mode transition acknowledged\n")
	} else {
		fmt.Fprintf(log, "[Enter Inspect Command]: Could not enter inspect mode\n")
	}

	return res[0] == globals.CMD_ENTER_INSPECT
}

// ClearAnalogSDCommand clears the analog SD card data
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

// switch back to normal operation, and send the launch flag
// not sure how to get an ack here, since removing the launch flag means that the
// radio and everything else would be saturated already
func EnterLaunchMode(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Remove Radio Delay]: Requesting return to normal mode\n")
	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Remove Radio Delay]: Failed to enter normal mode\n")
		return false
	}

	fmt.Fprintf(log, "[Remove Radio Delay]: sending command to remove all delays\n")

	cmd := getDispatchCommand(globals.CMD_ENTER_LAUNCH_MODE)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Remove Radio Delay]: Read timed out")
		return false
	}

	return res[0] == globals.CMD_ENTER_LAUNCH_MODE
}

// to verify that the SD card is working
// 1. Check the current file size and the timestamp
// 2. Return the system back to normal mode and wait for 10s
// 3. Send the system back into inspect mode
// 4. Check the new file size and timestamp, it should be bigger than the previous one
type sdUpdate struct {
	FileSize      uint32
	LastTimestamp uint32
}

func getSDUpdate(conn SerialReaderWriter, log io.Writer, command byte) *sdUpdate {
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
	var updateData sdUpdate
	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
		return nil
	}
	fmt.Fprintf(log, "[SD Update]: file size: %d, last update timestamp: %d\n", updateData.FileSize, updateData.LastTimestamp)

	return &updateData
}

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
