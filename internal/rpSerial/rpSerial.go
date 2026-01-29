/**
Wrapper around the regular serial package to simplify the interface

This wrapper should:
1. Be able to list all the open ports and connect to one
2. Process and stream the incoming telemetry data
3. Send commands through the serial port
*/

package rpSerial

import (
	"bytes"

	"go.bug.st/serial"
	"go.uber.org/zap"
)

const TEMP_BUF_SIZE = 256
const STOP_SEQUENCE_SIZE = 2

type RpSerial struct {
	serial.Port

	logger       *zap.Logger
	stopSequence [STOP_SEQUENCE_SIZE]byte
	tempBuf      [TEMP_BUF_SIZE]byte
	tempBufIdx   int
}

func NewRPSerial(portName string, baudrate int, logger *zap.Logger) *RpSerial {
	mode := &serial.Mode{
		BaudRate: baudrate,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		logger.Fatal("Error opening serial port", zap.Error(err), zap.String("portName", portName))
	}

	return &RpSerial{
		Port:         port,
		logger:       logger,
		stopSequence: [STOP_SEQUENCE_SIZE]byte{'\r', '\n'},
		tempBufIdx:   0,
	}
}

func (r *RpSerial) sync() {
	r.logger.Warn("Resyncing serial port")
	twoBytes := [2]byte{0x0, 0x0}
	oneByte := [1]byte{}

	for !bytes.Equal(twoBytes[:], r.stopSequence[:]) {
		_, err := r.Read(oneByte[:])
		if err != nil {
			r.logger.Warn("Error while resyncing serial port", zap.Error(err))
		}

		// update the two byte sequence
		twoBytes[0] = twoBytes[1]
		twoBytes[1] = oneByte[0]
	}
}

// read a single message in the buffer, continue until you reach the end characters \r\n
func (r *RpSerial) ReadSingleMessage() []byte {
	tempBufIdx := 0
	tempBuf := [TEMP_BUF_SIZE]byte{}

	for {
		_, err := r.Read(tempBuf[tempBufIdx : tempBufIdx+1])

		if err != nil {
			r.logger.Error("Error while trying to read new sequence", zap.Error(err))
			r.sync()
			tempBufIdx = 0 // Reset on error/sync
			continue
		}
		tempBufIdx++

		if tempBufIdx >= 2 && tempBuf[tempBufIdx-2] == '\r' && tempBuf[tempBufIdx-1] == '\n' {
			break
		}

		// handle overflow
		if tempBufIdx >= TEMP_BUF_SIZE {
			r.logger.Warn("Buffer overflow, forcefully terminating")
			tempBuf[TEMP_BUF_SIZE-1] = 0
			return tempBuf[:]
		}
	}

	return tempBuf[:tempBufIdx-2]
}

func (r *RpSerial) WriteSingleMessage(message []byte, size int) {
	n, err := r.Write(message[:size])

	if err != nil {
		r.logger.Error("Error while trying to send message", zap.Error(err))
	} else {
		r.logger.Info("Wrote message to serial port", zap.Int("bytesWritten", n), zap.ByteString("string", message))
	}
}

func ListPorts() ([]string, error) {
	return serial.GetPortsList()
}
