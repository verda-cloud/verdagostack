package wizard

import (
	"reflect"
	"strings"

	"charm.land/lipgloss/v2"
)

// HintBarView renders contextual keyboard hints based on the current
// step's PromptType. It subscribes to StepChangedMsg and updates
// automatically as the wizard progresses.
type HintBarView struct {
	promptType PromptType
	style      lipgloss.Style
	sepStyle   lipgloss.Style
}

// NewHintBarView creates a HintBarView.
func NewHintBarView() *HintBarView {
	return &HintBarView{
		promptType: -1, // no step yet
		style:      lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		sepStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
}

// Update handles StepChangedMsg to update the hint based on PromptType.
func (v *HintBarView) Update(msg any) (string, []any) {
	if m, ok := msg.(StepChangedMsg); ok {
		v.promptType = m.PromptType
	}
	return v.render(), nil
}

// Subscribe limits this view to StepChangedMsg only.
func (v *HintBarView) Subscribe() []reflect.Type {
	return []reflect.Type{reflect.TypeOf(StepChangedMsg{})}
}

func (v *HintBarView) render() string {
	hints := v.hintsForPrompt(v.promptType)
	if len(hints) == 0 {
		return ""
	}
	sep := v.sepStyle.Render(" · ")
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = v.style.Render(h)
	}
	return "  " + strings.Join(parts, sep)
}

func (v *HintBarView) hintsForPrompt(pt PromptType) []string {
	switch pt {
	case SelectPrompt:
		return []string{"↑/↓ navigate", "type to filter", "enter select", "esc back"}
	case MultiSelectPrompt:
		return []string{"↑/↓ navigate", "space toggle", "enter confirm", "esc back"}
	case TextInputPrompt:
		return []string{"enter submit", "esc cancel"}
	case ConfirmPrompt:
		return []string{"y/n", "enter confirm"}
	case PasswordPrompt:
		return []string{"enter submit", "esc cancel"}
	default:
		return nil
	}
}
