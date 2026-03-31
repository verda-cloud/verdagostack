package bubbletea

import "github.com/charmbracelet/lipgloss"

// Shared styles for all bubbletea prompts.
var (
	promptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // green bold "?"
	titleStyle    = lipgloss.NewStyle().Bold(true)                                  // bold prompt text
	answerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))            // cyan for final answer
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))            // cyan ">"
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))            // cyan for selected item
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))             // dim for unselected
	checkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))            // green for [x]
	uncheckStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))             // dim for [ ]
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))             // red for errors
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))             // dim for hints
)
