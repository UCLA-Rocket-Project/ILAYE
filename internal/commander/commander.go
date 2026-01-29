package commander

import (
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const COMMAND_SEQUENCE_SIZE = 4
const COMMAND_BYTE_IDX = 2

// use non printable characters for the start sequence
const COMMAND_START_SEQ_1 = 0x1A
const COMMAND_START_SEQ_2 = 0x1B
const COMMAND_END_SEQ = 0x1C
const COMMAND_ACK_SEQ = 0xFF

const SD_CARD_TEST_TIMEOUT = 10 * time.Second

type SerialReaderWriter interface {
	WriteSingleMessage(message []byte, size int)
	ReadSingleMessage() []byte
}

func getDispatchCommand(cmd byte) [COMMAND_SEQUENCE_SIZE]byte {
	return [COMMAND_SEQUENCE_SIZE]byte{COMMAND_START_SEQ_1, COMMAND_START_SEQ_2, cmd, COMMAND_END_SEQ}
}

// need some sort of verification for the commands
func EnterNormalCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Enter Normal Command]: sending command to enter normal mode...\n")

	cmd := getDispatchCommand(globals.CMD_ENTER_NORMAL)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res := conn.ReadSingleMessage()

	if res[0] == globals.CMD_ENTER_NORMAL {
		fmt.Fprintf(log, "[Enter Normal Command]: Normal mode transition acknowledged\n")
	} else {
		fmt.Fprintf(log, "[Enter Normal Command]: Could not enter normal mode")
	}

	return res[0] == globals.CMD_ENTER_NORMAL
}

// need some more verification for this
func EnterInspectCommand(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Enter Inspect Command]: sending command to enter inspect mode...\n")

	cmd := getDispatchCommand(globals.CMD_ENTER_INSPECT)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res := conn.ReadSingleMessage()

	if res[0] == globals.CMD_ENTER_INSPECT {
		fmt.Fprintf(log, "[Enter Inspect Command]: Inspect mode transition acknowledged\n")
	} else {
		fmt.Fprintf(log, "[Enter Inspect Command]: Could not enter inspect mode\n")
	}

	return res[0] == globals.CMD_ENTER_INSPECT
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

func getSDUpdate(conn SerialReaderWriter, log io.Writer) *sdUpdate {
	sdUpdateMessage := getDispatchCommand(globals.CMD_GET_ANALOG_SD_UPDATE)
	conn.WriteSingleMessage(sdUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[SD Update]: Sent command requesting SD card update\n")

	res := conn.ReadSingleMessage()
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
	fmt.Fprintf(log, "[Check Analog SD]: Dispatching sd card checker...\n")
	firstUpdate := getSDUpdate(conn, log)

	fmt.Fprintf(log, "[Check Analog SD]: Entering normal mode...\n")
	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Check Analog SD]: Failed to enter normal mode\n")
		return false
	}

	time.Sleep(SD_CARD_TEST_TIMEOUT)
	fmt.Fprintf(log, "[Check Analog SD]: Entering inspect mode...\n")
	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Check Analog SD]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Check Analog SD]: Dispatching sd card checker again...\n")
	secondUpdate := getSDUpdate(conn, log)

	return firstUpdate.FileSize < secondUpdate.FileSize && firstUpdate.LastTimestamp < secondUpdate.LastTimestamp
}
