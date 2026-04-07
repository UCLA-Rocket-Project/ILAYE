package commander

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// NOTE: we are removing this because we no longer have LCs on the flight board
// func CheckAnalogLCCommand(conn SerialReaderWriter, log io.Writer) bool {
// 	fmt.Fprintf(log, "[Check Analog LC]: Entering inspect mode\n")
// 	if !EnterInspectCommand(conn, log) {
// 		fmt.Fprintf(log, "[Check Analog LC]: Failed to enter inspect mode\n")
// 		return false
// 	}

// 	sdUpdateMessage := getDispatchCommand(globals.CMD_GET_ANALOG_LC_READING)
// 	conn.WriteSingleMessage(sdUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
// 	fmt.Fprintf(log, "[Chceck Analog LC]: Sent command requesting LC update\n")

// 	res, err := conn.ReadSingleOrTimeout()
// 	if err != nil {
// 		fmt.Fprintf(log, "[Check Analog LC]: Read timed out")
// 		return false
// 	}
// 	fmt.Fprintf(log, "[Check Analog LC]: Receieved response from boards\n")
// 	streamReader := bytes.NewReader(res[:])
// 	var updateData float32
// 	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
// 		return false
// 	}
// 	fmt.Fprintf(log, "[Check Analog LC]: Raw Reading: %f, Calibrated Reading: %f\n", updateData, -223810.211*updateData+10.86155)

// 	return true
// }

type ptUpdate struct {
	Ch [3]float32
}

func CheckAnalogPTCommand(conn SerialReaderWriter, log io.Writer, boardType string, command byte) bool {
	fmt.Fprintf(log, "[Check %s PTs]: Entering inspect mode\n", boardType)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check %s PTs]: Failed to enter inspect mode\n", boardType)
		return false
	}

	ptRequestMessage := getDispatchCommand(command)
	conn.WriteSingleMessage(ptRequestMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[Chceck %s PTs]: Sent command requesting PTs update\n", boardType)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Check %s PTs]: Read timed out", boardType)
		return false
	}
	fmt.Fprintf(log, "[Check %s PTs]: Receieved response from boards\n", boardType)
	streamReader := bytes.NewReader(res[:])
	var updateData ptUpdate
	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
		return false
	}
	fmt.Fprintf(
		log,
		"[Check %s PTs]: PT raw readings: %f %f %f\nCalibrated readings: %f %f %f\n",
		boardType,
		updateData.Ch[0],
		updateData.Ch[1],
		updateData.Ch[2],
		updateData.Ch[0],
		updateData.Ch[1],
		updateData.Ch[2],
	)

	return true
}
