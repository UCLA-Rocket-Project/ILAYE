package commander

import (
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
