package terminal

import (
	"github.com/charmbracelet/lipgloss"
)

// Vercel-inspired color palette
var (
	// Base colors
	colorBg        = lipgloss.Color("#000000")
	colorFg        = lipgloss.Color("#EDEDED")
	colorMuted     = lipgloss.Color("#666666")
	colorBorder    = lipgloss.Color("#333333")
	colorHighlight = lipgloss.Color("#0070F3") // Vercel blue

	// Status colors
	colorSuccess = lipgloss.Color("#50E3C2") // Teal/cyan for success
	colorError   = lipgloss.Color("#E00")    // Red for errors
	colorRunning = lipgloss.Color("#0070F3") // Blue for running
	colorPending = lipgloss.Color("#666666") // Gray for pending
	colorWarning = lipgloss.Color("#F5A623") // Orange for warnings
)

// Layout styles
var (
	// Main container with border
	containerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	// Header style (like "Build Logs" header)
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorFg).
			Padding(0, 1)

	// Subheader/muted text
	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Footer hint style
	hintStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)
)

// List styles
var (
	// Selected item in a list
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorHighlight).
				Bold(true)

	// Unselected item in a list
	normalItemStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	// Cursor pointer
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)
)

// Status indicator styles
var (
	pendingStyle = lipgloss.NewStyle().
			Foreground(colorPending)

	runningStyle = lipgloss.NewStyle().
			Foreground(colorRunning).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)
)

// Log styles
var (
	// Timestamp style for logs (like 12:14:20.141)
	timestampStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Log content style
	logContentStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	// Indented log line
	logIndent = "   "
)

// Status icons
const (
	iconPending = "○"
	iconSuccess = "✓"
	iconFail    = "✕"
	// Spinner frames are handled by the spinner component
)

// Helper functions
func renderCursor(active bool) string {
	if active {
		return cursorStyle.Render("▸")
	}
	return " "
}

func renderCheckbox(checked bool) string {
	if checked {
		return successStyle.Render("[✓]")
	}
	return mutedStyle.Render("[ ]")
}

func renderStatusIcon(status TestStatus, spinnerView string) string {
	switch status {
	case StatusPending:
		return pendingStyle.Render(iconPending)
	case StatusRunning:
		return runningStyle.Render(spinnerView)
	case StatusPass:
		return successStyle.Render(iconSuccess)
	case StatusFail:
		return errorStyle.Render(iconFail)
	default:
		return pendingStyle.Render(iconPending)
	}
}

func renderTestName(name string, status TestStatus) string {
	switch status {
	case StatusPass:
		return successStyle.Render(name)
	case StatusFail:
		return errorStyle.Render(name)
	case StatusRunning:
		return runningStyle.Render(name)
	default:
		return mutedStyle.Render(name)
	}
}

func renderHint(text string) string {
	return hintStyle.Render(text)
}
