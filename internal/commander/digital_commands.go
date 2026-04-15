package commander

import (
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type ShockAccelNum byte

type shockData struct {
	AccX float32
	AccY float32
	AccZ float32
}

func CheckDigitalShockCmd(conn SerialReaderWriter, log io.Writer, digitalBoardVersion string, command byte) bool {
	var shockNum int

	switch command {
	case globals.CMD_GET_DIGITAL_V1_SHOCK_1_READING:
		shockNum = 1
	case globals.CMD_GET_DIGITAL_V2_SHOCK_1_READING:
		shockNum = 1
	case globals.CMD_GET_DIGITAL_V2_SHOCK_2_READING:
		shockNum = 2
	}

	fmt.Fprintf(log, "[Check Digital %s Shock %d]: Entering inspect mode\n", digitalBoardVersion, shockNum)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital %s Shock %d]: Failed to enter inspect mode\n", digitalBoardVersion, shockNum)
		return false
	}

	shockUpdateMessage := getDispatchCommand(command)
	conn.WriteSingleMessage(shockUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[Check Digital %s Shock %d]: Sent command requesting Shock %d update\n", digitalBoardVersion, shockNum, shockNum)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Check Digital %s Shock %d]: Read timed out", digitalBoardVersion, shockNum)
		return false
	}
	fmt.Fprintf(log, "[Check Digital %s Shock %d]: Receieved response from boards of len %d\n", digitalBoardVersion, shockNum, len(res))
	streamReader := bytes.NewReader(res[:])
	var shockData shockData
	if err := binary.Read(streamReader, binary.LittleEndian, &shockData); err != nil {
		fmt.Fprintf(log, "[Check Digital %s Shock %d]: Error decoding shock response\n", digitalBoardVersion, shockNum)
		return false
	}

	fmt.Fprintf(
		log,
		"[Check Digital %s Shock %d]: \nShock data: %f, %f, %f\n",
		digitalBoardVersion, shockNum, shockData.AccX, shockData.AccY, shockData.AccZ,
	)
	return true
}

type IMUData struct {
	AccX float32
	AccY float32
	AccZ float32

	GyrX float32
	GyrY float32
	GyrZ float32

	Timestamp uint32
}

func CheckDigitalIMUCommand(conn SerialReaderWriter, log io.Writer, digitalBoardVersion string, command byte) bool {
	fmt.Fprintf(log, "[Check Digital %s IMU]: Entering inspect mode\n", digitalBoardVersion)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital %s IMU]: Failed to enter inspect mode\n", digitalBoardVersion)
		return false
	}

	sdUpdateMessage := getDispatchCommand(command)
	conn.WriteSingleMessage(sdUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[Chceck Digital %s IMU]: Sent command requesting IMU update\n", digitalBoardVersion)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Check Digital %s IMU]: Read timed out", digitalBoardVersion)
		return false
	}
	fmt.Fprintf(log, "[Check Digital %s IMU]: Receieved response from boards\n", digitalBoardVersion)
	streamReader := bytes.NewReader(res[:])
	var updateData IMUData
	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
		return false
	}
	fmt.Fprintf(
		log,
		"[Check Digital %s IMU]: \nTimestamp: %d\nAccX %f, AccY %f, AccZ %f, \nGyrX %f, GyrY %f, GyrZ %f\n",
		digitalBoardVersion, updateData.Timestamp,
		updateData.AccX, updateData.AccY, updateData.AccZ,
		updateData.GyrX, updateData.GyrY, updateData.GyrZ,
	)

	return true
}

type AltimeterData struct {
	Temp      int32
	Pressure  int32
	Timestamp uint32
}

func CheckDigitalAltimeterCommand(conn SerialReaderWriter, log io.Writer, digitalBoardVersion string, command byte) bool {
	fmt.Fprintf(log, "[Check Digital %s Altimeter]: Entering inspect mode\n", digitalBoardVersion)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital %s Altimeter]: Failed to enter inspect mode\n", digitalBoardVersion)
		return false
	}

	sdUpdateMessage := getDispatchCommand(command)
	conn.WriteSingleMessage(sdUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[Chceck Digital %s Altimeter]: Sent command requesting Altimeter update\n", digitalBoardVersion)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Check Digital %s Altimeter]: Read timed out\n", digitalBoardVersion)
		return false
	}
	fmt.Fprintf(log, "[Check Digital %s Altimeter]: Receieved response from boards\n", digitalBoardVersion)
	streamReader := bytes.NewReader(res[:])
	var updateData AltimeterData
	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
		return false
	}
	fmt.Fprintf(
		log,
		"[Check Digital %s Altimeter]: \nTimestamp: %d\nTemp: %f, Pressure: %f",
		digitalBoardVersion, updateData.Timestamp,
		float32(updateData.Temp)/100, float32(updateData.Pressure)/100,
	)

	return true
}

type GPSData struct {
	Lat  int32
	Long int32
}

func CheckDigitalGPSCommand(conn SerialReaderWriter, log io.Writer, digitalBoardVersion string, command byte) bool {
	fmt.Fprintf(log, "[Check Digital %s GPS]: Entering inspect mode\n", digitalBoardVersion)
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Check Digital %s GPS]: Failed to enter inspect mode\n", digitalBoardVersion)
		return false
	}

	sdUpdateMessage := getDispatchCommand(command)
	conn.WriteSingleMessage(sdUpdateMessage[:], COMMAND_SEQUENCE_SIZE)
	fmt.Fprintf(log, "[Chceck Digital %s GPS]: Sent command requesting GPS update\n", digitalBoardVersion)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Check Digital %s GPS]: Read timed out\n", digitalBoardVersion)
		return false
	}
	fmt.Fprintf(log, "[Check Digital %s GPS]: Receieved response from boards\n", digitalBoardVersion)
	streamReader := bytes.NewReader(res[:])
	var updateData GPSData
	if err := binary.Read(streamReader, binary.LittleEndian, &updateData); err != nil {
		return false
	}
	fmt.Fprintf(
		log,
		"[Check Digital %s GPS]: Lat: %d Long: %d",
		digitalBoardVersion,
		updateData.Lat, updateData.Long,
	)

	return true
}
