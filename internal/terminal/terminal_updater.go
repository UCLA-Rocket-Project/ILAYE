package terminal

import (
	"UCLA-Rocket-Project/ILAYE/internal/commander"
	"UCLA-Rocket-Project/ILAYE/internal/globals"

	tea "github.com/charmbracelet/bubbletea"
)

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
