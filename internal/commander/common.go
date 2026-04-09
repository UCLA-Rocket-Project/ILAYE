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

const SD_CARD_TEST_TIMEOUT = 5 * time.Second

type SerialReaderWriter interface {
	WriteSingleMessage(message []byte, size int)
	ReadSingleOrTimeout() ([]byte, error)
}

func getDispatchCommand(cmd byte) [COMMAND_SEQUENCE_SIZE]byte {
	// for consistency with terminal, use carraige return when sending back a command
	return [COMMAND_SEQUENCE_SIZE]byte{cmd, '+', '+', '+'}
}

// to verify that the SD card is working
// 1. Check the current file size and the timestamp
// 2. Return the system back to normal mode and wait for 10s
// 3. Send the system back into inspect mode
// 4. Check the new file size and timestamp, it should be bigger than the previous one
type sdUpdate struct {
	LastTimestamp int64
	FileSize      uint32
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
		fmt.Fprintf(log, "[SD Update]: Could not decode board response, %s\n", err)
		return nil
	}
	fmt.Fprintf(log, "[SD Update]: file size: %d, last update timestamp: %d\n", updateData.FileSize, updateData.LastTimestamp)

	return &updateData
}

type jumpClockUplinkPaylod struct {
	CommandCode            uint8
	CurrentTimeStampMicros int64
}

func JumpClocks(conn SerialReaderWriter, log io.Writer) bool {
	// enter inspect mode first
	fmt.Fprintf(log, "[Jump Clock]: Entering inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Jump Clock]: Failed to enter inspect mode\n")
		return false
	}

	// Pre-allocate a slice of the exact size (e.g., 4 bytes for uint32 + 8 bytes for int64 = 12 bytes)
	// do this rather than use a struct so that there would be no padding,
	// since the receiver does not expect any padding
	// add 3 for the command sequence at the end
	messageBytes := make([]byte, 9+3)
	// Pack the bytes manually (using Little Endian here)
	clkMicro := time.Now().UnixMicro()
	messageBytes[0] = byte(globals.CMD_JUMP_CLK)
	binary.LittleEndian.PutUint64(messageBytes[1:9], uint64(clkMicro))
	copy(messageBytes[9:12], []byte("+++"))
	conn.WriteSingleMessage(messageBytes, len(messageBytes))

	fmt.Fprintf(log, "[Jump Clock]: Inspect mode transition success, sending command to jump clock to %d\n", clkMicro)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Jump Clock]: Read timed out")
		return false
	}

	fmt.Fprintf(log, "[Jump Clock]: Receieved response from boards of len %d\n", len(res))
	streamReader := bytes.NewReader(res[:])
	var boardClock int64
	if err := binary.Read(streamReader, binary.LittleEndian, &boardClock); err != nil {
		fmt.Fprintf(log, "[Jump Clock]: Error decoding clock response\n")
		return false
	}

	fmt.Fprintf(
		log,
		"[Jump Clock]: Radio board timestamp %d\n", boardClock,
	)
	return true
}

func InspectSDCards(conn SerialReaderWriter, log io.Writer, boardType string, command byte) bool {
	fmt.Fprintf(log, "[Check %s SD]: Entering inspect mode\n", boardType)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check %s SD]: Failed to enter inspect mode\n", boardType)
		return false
	}

	fmt.Fprintf(log, "[Check %s SD]: Dispatching sd card checker\n", boardType)
	firstUpdate := getSDUpdate(conn, log, command)

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
	secondUpdate := getSDUpdate(conn, log, command)

	if secondUpdate == nil {
		return false
	}

	return (firstUpdate.FileSize < secondUpdate.FileSize && firstUpdate.LastTimestamp < secondUpdate.LastTimestamp)
}

func ClearSDCard(conn SerialReaderWriter, log io.Writer, boardType string, command byte) bool {
	fmt.Fprintf(log, "[Clear %s SD]: Entering inspect mode\n", boardType)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Clear %s SD]: Failed to enter inspect mode\n", boardType)
		return false
	}

	fmt.Fprintf(log, "[Clear %s SD]: sending command to clear %s SD card\n", boardType, boardType)

	cmd := getDispatchCommand(command)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Clear %s SD]: Read timed out", boardType)
		return false
	}

	streamReader := bytes.NewReader(res[:])
	var freeSpace uint32
	if err := binary.Read(streamReader, binary.LittleEndian, &freeSpace); err != nil {
		fmt.Fprintf(log, "[Clear %s SD]: Could not clear %s SD card\n", boardType, boardType)
		return false
	}

	fmt.Fprintf(log, "[Clear %s SD]: Clear command acknowledged. Free space is now: %d MB\n", boardType, freeSpace)

	return true
}
