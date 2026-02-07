package globals

// all possible command sequences
const (
	CMD_ENTER_NORMAL          = 0x00
	CMD_ENTER_INSPECT         = 0x01
	CMD_GET_ANALOG_SD_UPDATE  = 0xA0
	CMD_GET_ANALOG_LC_READING = 0xA1
	CMD_CLEAR_ANALOG_SD       = 0xAE
	CMD_END_SEND_DELAY        = 0x04
	CMD_TIMEOUT               = 0xFF
)
