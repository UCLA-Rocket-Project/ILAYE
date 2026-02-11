package commander

import (
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"fmt"
	"io"
)

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

// switch back to normal operation, and send the launch flag
// not sure how to get an ack here, since removing the launch flag means that the
// radio and everything else would be saturated already
func EnterLaunchMode(conn SerialReaderWriter, log io.Writer) bool {
	fmt.Fprintf(log, "[Enter Launch Mode]: Requesting enter inspect mode\n")
	if !EnterInspectCommand(conn, log) {
		fmt.Fprintf(log, "[Enter Launch Mode]: Failed to enter inspect mode\n")
		return false
	}

	fmt.Fprintf(log, "[Enter Launch Mode]: sending command to remove all delays\n")

	cmd := getDispatchCommand(globals.CMD_ENTER_LAUNCH_MODE)
	conn.WriteSingleMessage(cmd[:], COMMAND_SEQUENCE_SIZE)

	res, err := conn.ReadSingleOrTimeout()
	if err != nil {
		fmt.Fprintf(log, "[Enter Launch Mode]: Read timed out")
		return false
	} else if res[0] != globals.CMD_ENTER_LAUNCH_MODE {
		fmt.Fprintf(log, "[Enter Launch Mode]: Could not enter launch mode")
	}

	fmt.Fprintf(log, "[Enter Launch Mode]: Final transition to normal mode")

	if !EnterNormalCommand(conn, log) {
		fmt.Fprintf(log, "[Enter Launch Mode]: Failed to enter normal mode\n")
		return false
	}

	return true
}
