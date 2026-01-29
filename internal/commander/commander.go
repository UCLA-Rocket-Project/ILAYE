package commander

// all possible command sequences
const (
	CMD_ENTER_NORMAL         = 0x00
	CMD_ENTER_INSPECT        = 0x01
	CMD_GET_ANALOG_SD_UPDATE = 0xA0
)

const COMMAND_SEQUENCE_SIZE = 4
const COMMAND_BYTE_IDX = 2

// use non printable characters for the start sequence
const COMMAND_START_SEQ_1 = 0x1A
const COMMAND_START_SEQ_2 = 0x1B
const COMMAND_END_SEQ = 0x1C

func GetDispatchCommand(cmd byte) [COMMAND_SEQUENCE_SIZE]byte {
	return [COMMAND_SEQUENCE_SIZE]byte{COMMAND_START_SEQ_1, COMMAND_START_SEQ_2, cmd, COMMAND_END_SEQ}
}
