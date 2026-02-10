package terminal

import (
	"UCLA-Rocket-Project/ILAYE/internal/commander"
	"UCLA-Rocket-Project/ILAYE/internal/globals"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case LogMsg:
		// Find currently running test and append log with timestamp
		for i := range m.results {
			if m.results[i].Status == StatusRunning {
				m.results[i].Logs = append(m.results[i].Logs, LogEntry{
					Timestamp: time.Now(),
					Content:   string(msg),
				})
				break
			}
		}
		return m, waitForLog(m.logChan)
	case TestStartMsg:
		if msg.Index >= 0 && msg.Index < len(m.results) {
			m.results[msg.Index].Status = StatusRunning
			m.results[msg.Index].StartTime = time.Now()
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
	case VIEW_SELECT_MODE:
		return m.updateSelectMode(msg)
	case VIEW_SELECT_TESTS:
		return m.updateSelectTests(msg)
	case VIEW_TEST_RUNNER:
		return m.updateTestRunner(msg)
	case VIEW_SELECT_COMMANDS:
		return m.updateSelectCommands(msg)
	case VIEW_COMMAND_RUNNER:
		return m.updateCommandRunner(msg)
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
			return m, tea.Batch(
				connectToPort(m.connector, m.potentialPorts[m.cursor]),
				m.spinner.Tick,
			)
		}
	}

	return m, nil
}

func (m model) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectionSuccessMsg:
		m.serial = msg
		m.cursor = 0
		m.uiState = VIEW_SELECT_MODE
		return m, nil
	case connectionErrorMsg:
		m.err = msg
		m.uiState = VIEW_LIST_PORTS
		return m, nil
	}
	return m, nil
}

func (m model) updateSelectMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(modeOptions)-1 {
				m.cursor++
			}
		case "enter":
			selectedMode := m.cursor
			m.cursor = 0
			if selectedMode == 0 {
				// Run Tests selected
				m.uiState = VIEW_SELECT_TESTS
			} else {
				// Run Commands selected
				m.uiState = VIEW_SELECT_COMMANDS
			}
			return m, nil
		}
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
		case "b":
			// Go back to mode selection
			m.uiState = VIEW_SELECT_MODE
			m.cursor = 0
			return m, nil
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
						Logs:   []LogEntry{},
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
					case globals.CMD_CLEAR_ANALOG_SD:
						success = commander.ClearAnalogSDCommand(m.serial, w)
					case globals.CMD_GET_ANALOG_SD_UPDATE:
						success = commander.CheckAnalogSDCommand(m.serial, w)
					case globals.CMD_GET_ANALOG_LC_READING:
						success = commander.CheckAnalogLCCommand(m.serial, w)
					}

					w.ch <- TestResultMsg{Index: resultIdx, Success: success}
					resultIdx++
				}
			}()

			return m, tea.Batch(waitForLog(m.logChan), m.spinner.Tick)
		}
	}

	return m, nil
}

func (m model) updateTestRunner(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if all tests are complete
	allComplete := true
	for _, res := range m.results {
		if res.Status == StatusPending || res.Status == StatusRunning {
			allComplete = false
			break
		}
	}

	// Auto-return to test selection when all tests are done
	if allComplete && len(m.results) > 0 {
		m.uiState = VIEW_SELECT_TESTS
		m.cursor = 0
		m.selectedTests = make(map[int]struct{})
		m.results = nil
		return m, nil
	}

	return m, nil
}

func (m model) updateSelectCommands(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(availableCommands)-1 {
				m.cursor++
			}
		case "b":
			// Go back to mode selection
			m.uiState = VIEW_SELECT_MODE
			m.cursor = 0
			return m, nil
		case " ":
			if m.cursor == 0 {
				if _, ok := m.selectedCommands[m.cursor]; !ok {
					for i := range len(availableCommands) {
						m.selectedCommands[i] = struct{}{}
					}
				} else {
					for i := range len(availableCommands) {
						delete(m.selectedCommands, i)
					}
				}
			} else {
				if _, ok := m.selectedCommands[m.cursor]; !ok {
					m.selectedCommands[m.cursor] = struct{}{}
				} else {
					delete(m.selectedCommands, m.cursor)
				}
			}
		case "enter":
			if len(m.selectedCommands) == 0 {
				return m, nil
			}
			m.uiState = VIEW_COMMAND_RUNNER
			m.cursor = 0
			m.logChan = make(chan any)

			// Initialize results
			m.results = []TestResult{}
			for idx, cmd := range availableCommands {
				if idx == 0 {
					continue
				}
				if _, ok := m.selectedCommands[idx]; ok {
					m.results = append(m.results, TestResult{
						Name:   cmd.commandName,
						Status: StatusPending,
						Logs:   []LogEntry{},
					})
				}
			}

			// Run commands in a separate goroutine
			go func() {
				defer close(m.logChan)
				w := &chanWriter{ch: m.logChan}
				resultIdx := 0
				for idx := range availableCommands {
					if idx == 0 {
						continue
					}
					if _, ok := m.selectedCommands[idx]; !ok {
						continue
					}

					w.ch <- TestStartMsg{Index: resultIdx}

					success := false
					switch availableCommands[idx].opCode {
					case globals.CMD_CLEAR_ANALOG_SD:
						success = commander.ClearAnalogSDCommand(m.serial, w)
					case globals.CMD_ENTER_LAUNCH_MODE:
						success = commander.EnterLaunchMode(m.serial, w)
					}

					w.ch <- TestResultMsg{Index: resultIdx, Success: success}
					resultIdx++
				}
			}()

			return m, tea.Batch(waitForLog(m.logChan), m.spinner.Tick)
		}
	}

	return m, nil
}

func (m model) updateCommandRunner(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if all commands are complete
	allComplete := true
	for _, res := range m.results {
		if res.Status == StatusPending || res.Status == StatusRunning {
			allComplete = false
			break
		}
	}

	// Auto-return to command selection when all commands are done
	if allComplete && len(m.results) > 0 {
		m.uiState = VIEW_SELECT_COMMANDS
		m.cursor = 0
		m.selectedCommands = make(map[int]struct{})
		m.results = nil
		return m, nil
	}

	return m, nil
}
