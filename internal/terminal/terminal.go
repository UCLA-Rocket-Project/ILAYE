package terminal

import (
	"UCLA-Rocket-Project/ILAYE/internal/commander"
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"fmt"
	"os"
	"strings"

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
	selectedTests map[int]struct{}

	// test runner internal state
	// test runner internal state
	results []TestResult
	logChan chan any
}

type TestStatus int

const (
	StatusPending TestStatus = iota
	StatusRunning
	StatusPass
	StatusFail
)

type TestResult struct {
	Name   string
	Logs   []string
	Status TestStatus
}

// satisfy the logging interface
type LogMsg string

type TestStartMsg struct {
	Index int
}

type TestResultMsg struct {
	Index   int
	Success bool
}

type chanWriter struct {
	ch chan any
}

func (w *chanWriter) Write(p []byte) (n int, err error) {
	w.ch <- LogMsg(string(p))
	return len(p), nil
}

func waitForLog(sub <-chan any) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-sub
		if !ok {
			return nil
		}
		return msg
	}
}

type commandAndDesc struct {
	commandName string
	opCode      byte
}

var logPool []string
var availableTests []commandAndDesc = []commandAndDesc{
	{"Select All", 0xFF},
	{"Enter Normal Mode", globals.CMD_ENTER_NORMAL},
	{"Enter Inspect Mode", globals.CMD_ENTER_INSPECT},
	{"Get Analog SD Card Update", globals.CMD_GET_ANALOG_SD_UPDATE},
}

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
		selectedTests:  make(map[int]struct{}),
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
	case LogMsg:
		// Find currently running test and append log
		for i := range m.results {
			if m.results[i].Status == StatusRunning {
				m.results[i].Logs = append(m.results[i].Logs, string(msg))
				break
			}
		}
		return m, waitForLog(m.logChan)
	case TestStartMsg:
		if msg.Index >= 0 && msg.Index < len(m.results) {
			m.results[msg.Index].Status = StatusRunning
		}
		return m, waitForLog(m.logChan)
	case TestResultMsg:
		if msg.Index >= 0 && msg.Index < len(m.results) {
			if msg.Success {
				m.results[msg.Index].Status = StatusPass
			} else {
				m.results[msg.Index].Status = StatusFail
			}
		}
		return m, waitForLog(m.logChan)
	}

	switch m.uiState {
	case VIEW_LIST_PORTS:
		return m.updatePortSelection(msg)
	case VIEW_LOADING:
		return m.updateLoading(msg)
	case VIEW_SELECT_TESTS:
		return m.updateSelectTests(msg)
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

func (m model) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectionSuccessMsg:
		m.serial = msg
		m.cursor = 0
		m.uiState = VIEW_SELECT_TESTS
		return m, nil
	case connectionErrorMsg:
		m.err = msg
		m.uiState = VIEW_LIST_PORTS
		return m, nil
	}
	return m, nil
}

func (m model) updateSelectTests(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(availableTests)-1 {
				m.cursor++
			}
		case " ":
			if m.cursor == 0 {
				if _, ok := m.selectedTests[m.cursor]; !ok {
					for i := range len(availableTests) {
						m.selectedTests[i] = struct{}{}
					}
				} else {
					for i := range len(availableTests) {
						delete(m.selectedTests, i)
					}
				}

			} else {
				if _, ok := m.selectedTests[m.cursor]; !ok {
					m.selectedTests[m.cursor] = struct{}{}
				} else {
					delete(m.selectedTests, m.cursor)
				}
			}

		case "enter":
			if len(m.selectedTests) == 0 {
				// TODO: put some kind of error message here that says you have to select at least one test
				return m, nil
			}
			m.uiState = VIEW_TEST_RUNNER
			m.cursor = 0
			m.logChan = make(chan any)

			// Initialize results
			m.results = []TestResult{}
			// Calculate mapping from result index to availableTests index is not needed
			// if we iterate availableTests in order in both places.
			// But the goroutine iterates availableTests and only creates results for selected ones.
			// So we need to match that.
			for idx, test := range availableTests {
				if idx == 0 {
					continue
				}
				if _, ok := m.selectedTests[idx]; ok {
					m.results = append(m.results, TestResult{
						Name:   test.commandName,
						Status: StatusPending,
						Logs:   []string{},
					})
				}
			}

			// Run tests in a separate goroutine
			go func() {
				defer close(m.logChan)
				w := &chanWriter{ch: m.logChan}
				resultIdx := 0
				for idx := range availableTests { // iterate in order
					if idx == 0 {
						continue // skip select all
					}
					if _, ok := m.selectedTests[idx]; !ok {
						continue
					}

					w.ch <- TestStartMsg{Index: resultIdx}

					success := false
					// Map index to commander function
					switch availableTests[idx].opCode {
					case globals.CMD_ENTER_NORMAL:
						success = commander.EnterNormalCommand(m.serial, w)
					case globals.CMD_ENTER_INSPECT:
						success = commander.EnterInspectCommand(m.serial, w)
					case globals.CMD_GET_ANALOG_SD_UPDATE:
						success = commander.CheckAnalogSDCommand(m.serial, w)
					}

					w.ch <- TestResultMsg{Index: resultIdx, Success: success}
					resultIdx++
				}
			}()

			return m, waitForLog(m.logChan)
		}
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
		s += "See the tests available!\n\n"
		for i, test := range availableTests {
			if m.cursor == i {
				s += ">"
			} else {
				s += " "
			}
			s += " ["
			if _, ok := m.selectedTests[i]; ok {
				s += "x"
			} else {
				s += " "
			}
			s += fmt.Sprintf("] %s\n", test.commandName)
		}

		s += "\n\n<space> to select | <enter> to proceed\n"

	case VIEW_TEST_RUNNER:
		s += "Running Tests...\n\n"
		for _, res := range m.results {
			switch res.Status {
			case StatusPending:
				s += fmt.Sprintf("[ ] %s\n", res.Name)
			case StatusRunning:
				s += fmt.Sprintf("[...] %s\n", res.Name)
			case StatusPass:
				s += fmt.Sprintf("[✓] %s\n", res.Name)
			case StatusFail:
				s += fmt.Sprintf("[x] %s\n", res.Name)
			}

			if len(res.Logs) > 0 {
				for _, log := range res.Logs {
					s += fmt.Sprintf("    %s", log)
					if !strings.HasSuffix(log, "\n") {
						s += "\n"
					}
				}
			}
		}
	}

	return s
}
