package terminal

type UIState int

const (
	VIEW_LIST_PORTS UIState = iota
	VIEW_CONNECT_TO_PORT
	VIEW_SELECT_TESTS
	VIEW_TEST_RUNNER
)

// defines the internal state of the TUI
type model struct {
	// methods

	// global internal state
	uiState UIState
	cursor  int

	// connect to port internal state
	portName string

	// select tests internal state
	selectedTests map[int]byte

	// test runner internal state

}
