package bubbletea

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme defines the color palette for all TUI prompts.
type Theme struct {
	Accent  color.Color // cursor, selected items, final answers
	Success color.Color // prompt "?", checkmarks
	Error   color.Color // validation errors
	Dim     color.Color // unselected items, hints
}

// Built-in themes.
var (
	// ThemeDefault uses standard terminal ANSI colors.
	ThemeDefault = Theme{
		Accent:  lipgloss.Color("14"), // cyan
		Success: lipgloss.Color("10"), // green
		Error:   lipgloss.Color("9"),  // red
		Dim:     lipgloss.Color("8"),  // gray
	}

	// ThemeDracula uses the Dracula color palette.
	ThemeDracula = Theme{
		Accent:  lipgloss.Color("#bd93f9"), // purple
		Success: lipgloss.Color("#50fa7b"), // green
		Error:   lipgloss.Color("#ff5555"), // red
		Dim:     lipgloss.Color("#6272a4"), // comment
	}

	// ThemeCatppuccin uses the Catppuccin Mocha palette.
	ThemeCatppuccin = Theme{
		Accent:  lipgloss.Color("#89b4fa"), // blue
		Success: lipgloss.Color("#a6e3a1"), // green
		Error:   lipgloss.Color("#f38ba8"), // red
		Dim:     lipgloss.Color("#6c7086"), // overlay0
	}

	// ThemeNord uses the Nord color palette.
	ThemeNord = Theme{
		Accent:  lipgloss.Color("#88c0d0"), // frost
		Success: lipgloss.Color("#a3be8c"), // green
		Error:   lipgloss.Color("#bf616a"), // red
		Dim:     lipgloss.Color("#4c566a"), // polar night
	}

	// ThemeTokyoNight uses the Tokyo Night palette.
	ThemeTokyoNight = Theme{
		Accent:  lipgloss.Color("#7aa2f7"), // blue
		Success: lipgloss.Color("#9ece6a"), // green
		Error:   lipgloss.Color("#f7768e"), // red
		Dim:     lipgloss.Color("#565f89"), // comment
	}
)

// activeTheme is the current theme. Defaults to ThemeDefault.
var activeTheme = ThemeDefault

// SetTheme changes the color theme for all TUI prompts.
// Call this before creating any prompts (typically in main or init).
func SetTheme(t Theme) {
	activeTheme = t
	applyTheme()
}

// Shared styles derived from the active theme.
var (
	promptStyle   lipgloss.Style
	titleStyle    lipgloss.Style
	answerStyle   lipgloss.Style
	cursorStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	dimStyle      lipgloss.Style
	checkStyle    lipgloss.Style
	uncheckStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	hintStyle     lipgloss.Style
)

func init() {
	applyTheme()
}

func applyTheme() {
	t := activeTheme
	promptStyle = lipgloss.NewStyle().Foreground(t.Success).Bold(true)
	titleStyle = lipgloss.NewStyle().Bold(true)
	answerStyle = lipgloss.NewStyle().Foreground(t.Accent)
	cursorStyle = lipgloss.NewStyle().Foreground(t.Accent)
	selectedStyle = lipgloss.NewStyle().Foreground(t.Accent)
	dimStyle = lipgloss.NewStyle().Foreground(t.Dim)
	checkStyle = lipgloss.NewStyle().Foreground(t.Success)
	uncheckStyle = lipgloss.NewStyle().Foreground(t.Dim)
	errorStyle = lipgloss.NewStyle().Foreground(t.Error)
	hintStyle = lipgloss.NewStyle().Foreground(t.Dim)
}
