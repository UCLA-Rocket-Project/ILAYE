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
	"errors"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.uber.org/zap"
)

const TEMP_BUF_SIZE = 256

var startSequence = []byte{0x08, 0x09}
var stopSequence = []byte{0x1F, '\n'}

type RpSerial struct {
	serial.Port

	logger *zap.Logger
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
		Port:   port,
		logger: logger,
	}
}

func (r *RpSerial) Sync() {
	r.logger.Warn("Resyncing serial port")
	twoBytes := [2]byte{0x0, 0x0}
	oneByte := [1]byte{}

	for !bytes.Equal(twoBytes[:], stopSequence[:]) {
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

	// clear the serial terminal of anything that does not have to do with the actual data
	r.readTillStartSequence()

	for {
		_, err := r.Read(tempBuf[tempBufIdx : tempBufIdx+1])
		r.logger.Info("Got some byte at actual read ", zap.Int("receivedByte", int(tempBuf[tempBufIdx])))

		if err != nil {
			r.logger.Error("Error while trying to read new sequence", zap.Error(err))
			r.Sync()
			tempBufIdx = 0 // Reset on error/sync
			continue
		}
		tempBufIdx++

		if tempBufIdx >= 2 && tempBuf[tempBufIdx-2] == stopSequence[0] && tempBuf[tempBufIdx-1] == stopSequence[1] {
			break
		}

		// handle overflow
		if tempBufIdx >= TEMP_BUF_SIZE {
			tempBuf[TEMP_BUF_SIZE-1] = 0
			r.logger.Warn("Buffer overflow, forcefully terminating", zap.ByteString("currentContents", tempBuf[:]))
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
	// there is an issue of there being duplicate ports
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	// Use a map to track unique ports and filter out tty.* duplicates
	// On macOS, we prefer cu.* over tty.* for the same device
	uniquePorts := make(map[string]bool)
	filteredPorts := make([]string, 0)

	for _, port := range ports {
		// Skip tty.* ports if we're on macOS (they're duplicates of cu.* ports)
		if strings.Contains(port, "/dev/tty.") {
			continue
		}

		// For Windows COM ports and macOS cu.* ports, add them
		if !uniquePorts[port] {
			filteredPorts = append(filteredPorts, port)
			uniquePorts[port] = true
		}
	}

	// Separate into USB ports and other ports
	usbPorts := make([]string, 0)
	otherPorts := make([]string, 0)

	for _, port := range filteredPorts {
		lowerPort := strings.ToLower(port)
		// Check for USB serial ports
		if strings.Contains(lowerPort, "usbserial") || strings.Contains(lowerPort, "usb") {
			usbPorts = append(usbPorts, port)
		} else {
			otherPorts = append(otherPorts, port)
		}
	}

	// Return USB ports first, then other ports
	return append(usbPorts, otherPorts...), nil
}

func (r *RpSerial) readTillStartSequence() {
	tempBuf := [2]byte{0x00, 0x00}
	singleBuf := [1]byte{0x00}

	for {
		_, err := r.Read(singleBuf[:])

		r.logger.Info("Got some bytes while clearing for right start sequence ", zap.Int("receivedByte", int(singleBuf[0])))
		if err != nil {
			r.logger.Error("Error while trying to read new sequence", zap.Error(err))
			r.Sync()
			continue
		}

		tempBuf[0], tempBuf[1] = tempBuf[1], singleBuf[0]
		if tempBuf[0] == startSequence[0] && tempBuf[1] == startSequence[1] {
			return
		}
	}
}

func (r *RpSerial) ReadSingleOrTimeout() ([]byte, error) {
	resultCh := make(chan []byte, 1)

	go func() {
		resultCh <- r.ReadSingleMessage()
	}()

	select {
	case res := <-resultCh:
		return res, nil
	// timeout on the boards is 10 seconds
	case <-time.After(15 * time.Second):
		return nil, errors.New("read timeout")
	}
}
