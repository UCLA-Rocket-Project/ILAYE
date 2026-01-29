package terminal

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"UCLA-Rocket-Project/ILAYE/internal/commander"
)

// SerialConnection defines the methods required from the serial package
type SerialConnection interface {
	ReadSingleMessage() string
	WriteSingleMessage(message []byte, size int)
}

type PortLister func() ([]string, error)
type PortConnector func(port string) (SerialConnection, error)

type sessionState int

const (
	portSelectionView sessionState = iota
	connectingView                 // New state for loading
	testSelectionView
	runningTestView
)

// Messages
type connectionSuccessMsg SerialConnection
type connectionErrorMsg error

func connectToPort(connector PortConnector, port string) tea.Cmd {
	return func() tea.Msg {
		conn, err := connector(port)
		if err != nil {
			return connectionErrorMsg(err)
		}
		return connectionSuccessMsg(conn)
	}
}

type model struct {
	state sessionState

	// Dependencies
	listPorts func() ([]string, error)
	connect   func(string) (SerialConnection, error)
	serial    SerialConnection

	// State
	ports          []string
	availableTests []string
	cursor         int
	err            error
}

func initialModel(lister PortLister, connector PortConnector) *model {
	ports, err := lister()
	if err != nil {
		panic("Could not find open ports")
	}

	return &model{
		state:          portSelectionView,
		listPorts:      lister,
		connect:        connector,
		ports:          ports,
		availableTests: []string{"Get Analog Update", "Enter Normal Mode", "Enter Inspect Mode"},
		cursor:         0,
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	switch m.state {
	case portSelectionView:
		return m.updatePortSelection(msg)
	case connectingView:
		return m.updateConnecting(msg)
	case testSelectionView:
		return m.updateTestSelection(msg)
	case runningTestView:
		return m.updateRunningTest(msg)
	}
	return m, nil
}

func (m *model) updatePortSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.ports)-1 {
				m.cursor++
			}
		case "enter":
			selectedPort := m.ports[m.cursor]
			m.state = connectingView
			return m, connectToPort(m.connect, selectedPort)
		}
	}
	return m, nil
}

func (m *model) updateConnecting(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectionSuccessMsg:
		m.serial = msg
		m.state = testSelectionView
		m.cursor = 0
		return m, nil
	case connectionErrorMsg:
		m.err = msg
		m.state = portSelectionView
		return m, nil
	}
	return m, nil
}

func (m *model) updateTestSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.availableTests)-1 {
				m.cursor++
			}
		case "enter":
			m.state = runningTestView
			// Optionally trigger the test immediately e.g.
			// return m, m.runTestCmd(m.availableTests[m.cursor])
		case "esc":
			m.state = portSelectionView
			m.cursor = 0
			// m.serial.Close() ?? We don't have close in interface yet
		}
	}
	return m, nil
}

func (m *model) updateRunningTest(msg tea.Msg) (tea.Model, tea.Cmd) {
	// In running view, we just listen for keys to trigger commands
	// and potentially listen for incoming serial data (not implemented in this simplified version)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = testSelectionView
			return m, nil
		case "s":
			// Send the selected test command?
			// For now let's just use the hardcoded example but based on selection
			testName := m.availableTests[m.cursor]
			var cmd [4]byte
			switch testName {
			case "Get Analog Update":
				cmd = commander.GetDispatchCommand(commander.CMD_GET_ANALOG_SD_UPDATE)
			case "Enter Normal Mode":
				cmd = commander.GetDispatchCommand(commander.CMD_ENTER_NORMAL)
			case "Enter Inspect Mode":
				cmd = commander.GetDispatchCommand(commander.CMD_ENTER_INSPECT)
			}
			m.serial.WriteSingleMessage(cmd[:], commander.COMMAND_SEQUENCE_SIZE)
		}
	}
	return m, nil
}

func (m *model) View() string {
	s := ""
	if m.err != nil {
		s += fmt.Sprintf("Error: %v\n\n", m.err)
	}

	switch m.state {
	case portSelectionView:
		s += "Select a Port:\n\n"
		for i, port := range m.ports {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, port)
		}
	case connectingView:
		s += "Connecting...\n"
	case testSelectionView:
		s += "Select a Test:\n\n"
		for i, test := range m.availableTests {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, test)
		}
	case runningTestView:
		s += fmt.Sprintf("Running Test: %s\n\n", m.availableTests[m.cursor])
		s += "Press 's' to send command, 'esc' to go back.\n\n"
		s += "Log:\n"
	}

	return s
}

func Start(lister PortLister, connector PortConnector) {
	if _, err := tea.NewProgram(initialModel(lister, connector)).Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
