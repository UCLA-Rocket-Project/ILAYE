package terminal

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

type UIState int

const (
	VIEW_LIST_PORTS UIState = iota
	VIEW_SELECT_TESTS
	VIEW_TEST_RUNNER
	VIEW_LOADING
)

type SerialReaderWriter interface {
	WriteSingleMessage(message []byte, size int)
	ReadSingleMessage() []byte
}

type connectionSuccessMsg SerialReaderWriter
type connectionErrorMsg error

type PortLister func() ([]string, error)
type PortConnector func(string) (SerialReaderWriter, error)

// defines the internal state of the TUI
type model struct {
	// methods

	// global internal state
	uiState UIState
	cursor  int
	err     error

	// connect to port internal state
	potentialPorts []string
	portName       string
	connector      PortConnector
	serial         SerialReaderWriter

	// select tests internal state
	availableTests []string
	selectedTests  map[int]byte

	// test runner internal state
}

var logPool []string

func StartApplication(portLister PortLister, connector PortConnector, logger *zap.Logger) {
	if _, err := tea.NewProgram(initialModel(portLister, connector)).Run(); err != nil {
		logger.Fatal("Error starting TUI program", zap.Error(err))
		os.Exit(1)
	}
}

// TUI tries to use functional programming paradigms, so you return a new model everytime, rather
// then modify a pointer
func initialModel(portLister PortLister, connector PortConnector) model {
	ports, err := portLister()

	if err != nil {
		panic(fmt.Sprintf("Unable to open serial port: %v", err))
	}

	return model{
		uiState:        VIEW_LIST_PORTS,
		potentialPorts: ports,
		connector:      connector,
		availableTests: []string{
			"Select All",
			"Enter Normal Mode",
			"Enter Inspect Mode",
			"Get Analog SD Card Update",
		},
		selectedTests: make(map[int]byte),
	}
}

func connectToPort(connector PortConnector, port string) tea.Cmd {
	return func() tea.Msg {
		connection, err := connector(port)
		if err != nil {
			return connectionErrorMsg(err)
		}
		return connectionSuccessMsg(connection)
	}

}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	switch m.uiState {
	case VIEW_LIST_PORTS:
		return m.updatePortSelection(msg)
	case VIEW_LOADING:
		return m.updateLoading(msg)
	case VIEW_SELECT_TESTS:

	}

	return m, nil
}

func (m model) updatePortSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.potentialPorts)-1 {
				m.cursor++
			}
		case "enter":
			m.uiState = VIEW_LOADING
			return m, connectToPort(m.connector, m.potentialPorts[m.cursor])
		}
	}

	return m, nil
}

func (m *model) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectionSuccessMsg:
		m.serial = msg
		m.uiState = VIEW_SELECT_TESTS
		m.cursor = 0
		return m, nil
	case connectionErrorMsg:
		m.err = msg
		m.uiState = VIEW_LIST_PORTS
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	s := ""
	if m.err != nil {
		s += fmt.Sprintf("Error %v\n\n", m.err)
	}

	switch m.uiState {
	case VIEW_LIST_PORTS:
		s += "Select a port:\n\n"
		for i, port := range m.potentialPorts {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, port)
		}
	case VIEW_LOADING:
		s += "Connecting...\n"
	case VIEW_SELECT_TESTS:
		s += "Here are the tests that you see!\n"
	}

	return s
}
