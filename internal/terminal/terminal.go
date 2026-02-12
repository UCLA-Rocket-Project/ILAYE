package terminal

import (
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
)

type UIState int

const (
	VIEW_LIST_PORTS UIState = iota
	VIEW_SELECT_MODE
	VIEW_SELECT_TESTS
	VIEW_TEST_RUNNER
	VIEW_SELECT_COMMANDS
	VIEW_COMMAND_RUNNER
	VIEW_LOADING
)

type SerialReaderWriter interface {
	WriteSingleMessage(message []byte, size int)
	ReadSingleOrTimeout() ([]byte, error)
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

	// select commands internal state
	selectedCommands map[int]struct{}

	// test runner internal state
	results []TestResult
	logChan chan any

	// spinner for loading/running states
	spinner spinner.Model
}

type TestStatus int

const (
	StatusPending TestStatus = iota
	StatusRunning
	StatusPass
	StatusFail
)

type TestResult struct {
	Name      string
	Logs      []LogEntry
	Status    TestStatus
	StartTime time.Time
}

// LogEntry represents a single log line with its timestamp
type LogEntry struct {
	Timestamp time.Time
	Content   string
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
	// {"Enter Normal Mode", globals.CMD_ENTER_NORMAL},
	// {"Enter Inspect Mode", globals.CMD_ENTER_INSPECT},
	{"Get Radio SD Card Update", globals.CMD_GET_RADIO_SD_UPDATE},
	{"Get Analog SD Card Update", globals.CMD_GET_ANALOG_SD_UPDATE},
	{"Get Analog LC Reading", globals.CMD_GET_ANALOG_LC_READING},
	{"Get Digital SD Card Update", globals.CMD_GET_DIGITAL_SD_UPDATE},
	{"Get Digital Altimeter Reading", globals.CMD_GET_ALTIMETER_READING},
	{"Get Digital Shock 1 Reading", globals.CMD_GET_SHOCK_1_READING},
	{"Get Digital IMU Reading", globals.CMD_GET_IMU_READING},
}

var availableCommands []commandAndDesc = []commandAndDesc{
	{"Select All", 0xFF},
	{"Clear Radio SD", globals.CMD_CLEAR_RADIO_SD},
	{"Clear Analog SD", globals.CMD_CLEAR_ANALOG_SD},
	{"Clear Digital SD", globals.CMD_CLEAR_DIGITAL_SD},
	{"Prepare for launch (No coming back!)", globals.CMD_ENTER_LAUNCH_MODE},
}

var modeOptions = []string{"Run Tests", "Run Commands"}

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

	// Initialize spinner with a nice modern style
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = runningStyle

	return model{
		uiState:          VIEW_LIST_PORTS,
		potentialPorts:   ports,
		connector:        connector,
		selectedTests:    make(map[int]struct{}),
		selectedCommands: make(map[int]struct{}),
		spinner:          s,
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
	return m.spinner.Tick
}

func (m model) View() string {
	var s strings.Builder

	// Error display
	if m.err != nil {
		errBox := lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("⚠ Error: %v", m.err))
		s.WriteString(errBox + "\n\n")
	}

	switch m.uiState {
	case VIEW_LIST_PORTS:
		s.WriteString(m.viewPortSelection())
	case VIEW_LOADING:
		s.WriteString(m.viewLoading())
	case VIEW_SELECT_MODE:
		s.WriteString(m.viewSelectMode())
	case VIEW_SELECT_TESTS:
		s.WriteString(m.viewSelectTests())
	case VIEW_TEST_RUNNER:
		s.WriteString(m.viewTestRunner())
	case VIEW_SELECT_COMMANDS:
		s.WriteString(m.viewSelectCommands())
	case VIEW_COMMAND_RUNNER:
		s.WriteString(m.viewCommandRunner())
	}

	return s.String()
}

func (m model) viewPortSelection() string {
	var s strings.Builder

	// Header
	header := headerStyle.Render("▸ Select Serial Port")
	s.WriteString(header + "\n\n")

	// Port list
	for i, port := range m.potentialPorts {
		cursor := renderCursor(i == m.cursor)
		portName := port
		if i == m.cursor {
			portName = selectedItemStyle.Render(port)
		} else {
			portName = normalItemStyle.Render(port)
		}
		s.WriteString(fmt.Sprintf("  %s %s\n", cursor, portName))
	}

	s.WriteString("\n")
	s.WriteString(renderHint("  ↑/↓ navigate • enter select • q quit"))

	return s.String()
}

func (m model) viewLoading() string {
	var s strings.Builder

	// Spinner with connecting message
	spinnerView := m.spinner.View()
	s.WriteString(fmt.Sprintf("\n  %s  %s\n\n",
		spinnerView,
		normalItemStyle.Render("Connecting to port")))

	return s.String()
}

func (m model) viewSelectTests() string {
	var s strings.Builder

	// Header
	header := headerStyle.Render("▸ Select Tests to Run")
	s.WriteString(header + "\n\n")

	// Test list with checkboxes
	for i, test := range availableTests {
		cursor := renderCursor(i == m.cursor)
		_, isSelected := m.selectedTests[i]
		checkbox := renderCheckbox(isSelected)

		testName := test.commandName
		if i == m.cursor {
			testName = selectedItemStyle.Render(testName)
		} else if isSelected {
			testName = successStyle.Render(testName)
		} else {
			testName = normalItemStyle.Render(testName)
		}

		s.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, checkbox, testName))
	}

	s.WriteString("\n")
	s.WriteString(renderHint("  ↑/↓ navigate • space toggle • enter run • b back • q quit"))

	return s.String()
}

func (m model) viewTestRunner() string {
	var s strings.Builder

	// Count completed tests
	completed := 0
	passed := 0
	for _, res := range m.results {
		if res.Status == StatusPass || res.Status == StatusFail {
			completed++
			if res.Status == StatusPass {
				passed++
			}
		}
	}

	// Just add a small top margin
	s.WriteString("\n")

	// Test results
	for _, res := range m.results {
		// Status icon + test name
		icon := renderStatusIcon(res.Status, m.spinner.View())
		name := renderTestName(res.Name, res.Status)
		s.WriteString(fmt.Sprintf("  %s  %s\n", icon, name))

		// Log entries with timestamps
		if len(res.Logs) > 0 {
			for _, log := range res.Logs {
				timestamp := timestampStyle.Render(log.Timestamp.Format("15:04:05.000"))
				content := strings.TrimSuffix(log.Content, "\n")

				// Handle multi-line log content
				lines := strings.Split(content, "\n")
				for j, line := range lines {
					if j == 0 {
						s.WriteString(fmt.Sprintf("     %s  %s\n", timestamp, logContentStyle.Render(line)))
					} else {
						// Continuation lines without timestamp
						s.WriteString(fmt.Sprintf("     %s  %s\n", strings.Repeat(" ", 12), logContentStyle.Render(line)))
					}
				}
			}
		}
		s.WriteString("\n")
	}

	// Test summary
	if completed == len(m.results) && len(m.results) > 0 {
		// All tests done - show final summary
		summaryStyle := successStyle
		if passed != completed {
			summaryStyle = errorStyle
		}
		s.WriteString(mutedStyle.Render("  "+strings.Repeat("─", 40)) + "\n")
		s.WriteString(fmt.Sprintf("  %s\n\n", summaryStyle.Render(fmt.Sprintf("%d/%d tests passed", passed, len(m.results)))))
	}

	// Footer hint - show restart option when tests are done
	if completed == len(m.results) && len(m.results) > 0 {
		s.WriteString(renderHint("  r run more tests • q quit"))
	} else {
		s.WriteString(renderHint("  q quit"))
	}

	return s.String()
}

func (m model) viewSelectMode() string {
	var s strings.Builder

	// Header
	header := headerStyle.Render("▸ Select Mode")
	s.WriteString(header + "\n\n")

	// Mode options
	for i, option := range modeOptions {
		cursor := renderCursor(i == m.cursor)
		optionName := option
		if i == m.cursor {
			optionName = selectedItemStyle.Render(option)
		} else {
			optionName = normalItemStyle.Render(option)
		}
		s.WriteString(fmt.Sprintf("  %s %s\n", cursor, optionName))
	}

	s.WriteString("\n")
	s.WriteString(renderHint("  ↑/↓ navigate • enter select • q quit"))

	return s.String()
}

func (m model) viewSelectCommands() string {
	var s strings.Builder

	// Header
	header := headerStyle.Render("▸ Select Commands to Run")
	s.WriteString(header + "\n\n")

	// Command list with checkboxes
	for i, cmd := range availableCommands {
		cursor := renderCursor(i == m.cursor)
		_, isSelected := m.selectedCommands[i]
		checkbox := renderCheckbox(isSelected)

		cmdName := cmd.commandName
		if i == m.cursor {
			cmdName = selectedItemStyle.Render(cmdName)
		} else if isSelected {
			cmdName = successStyle.Render(cmdName)
		} else {
			cmdName = normalItemStyle.Render(cmdName)
		}

		s.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, checkbox, cmdName))
	}

	s.WriteString("\n")
	s.WriteString(renderHint("  ↑/↓ navigate • space toggle • enter run • b back • q quit"))

	return s.String()
}

func (m model) viewCommandRunner() string {
	var s strings.Builder

	// Count completed commands
	completed := 0
	succeeded := 0
	for _, res := range m.results {
		if res.Status == StatusPass || res.Status == StatusFail {
			completed++
			if res.Status == StatusPass {
				succeeded++
			}
		}
	}

	// Just add a small top margin
	s.WriteString("\n")

	// Command results
	for _, res := range m.results {
		// Status icon + command name
		icon := renderStatusIcon(res.Status, m.spinner.View())
		name := renderTestName(res.Name, res.Status)
		s.WriteString(fmt.Sprintf("  %s  %s\n", icon, name))

		// Log entries with timestamps
		if len(res.Logs) > 0 {
			for _, log := range res.Logs {
				timestamp := timestampStyle.Render(log.Timestamp.Format("15:04:05.000"))
				content := strings.TrimSuffix(log.Content, "\n")

				// Handle multi-line log content
				lines := strings.Split(content, "\n")
				for j, line := range lines {
					if j == 0 {
						s.WriteString(fmt.Sprintf("     %s  %s\n", timestamp, logContentStyle.Render(line)))
					} else {
						// Continuation lines without timestamp
						s.WriteString(fmt.Sprintf("     %s  %s\n", strings.Repeat(" ", 12), logContentStyle.Render(line)))
					}
				}
			}
		}
		s.WriteString("\n")
	}

	// Command summary
	if completed == len(m.results) && len(m.results) > 0 {
		// All commands done - show final summary
		summaryStyle := successStyle
		if succeeded != completed {
			summaryStyle = errorStyle
		}
		s.WriteString(mutedStyle.Render("  "+strings.Repeat("─", 40)) + "\n")
		s.WriteString(fmt.Sprintf("  %s\n\n", summaryStyle.Render(fmt.Sprintf("%d/%d commands succeeded", succeeded, len(m.results)))))
	}

	// Footer hint - show restart option when commands are done
	if completed == len(m.results) && len(m.results) > 0 {
		s.WriteString(renderHint("  r run more commands • b back to mode • q quit"))
	} else {
		s.WriteString(renderHint("  q quit"))
	}

	return s.String()
}
