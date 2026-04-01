package wizard

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressRegion displays a step progress bar.
// Responds to StepChangedMsg.
type ProgressRegion struct {
	last string
}

// NewProgressRegion creates a progress bar region.
func NewProgressRegion() *ProgressRegion {
	return &ProgressRegion{}
}

func (r *ProgressRegion) Update(msg any) (string, []any) {
	sc, ok := msg.(StepChangedMsg)
	if !ok {
		return r.last, nil
	}

	if sc.Total <= 1 {
		r.last = ""
		return r.last, nil
	}

	const barWidth = 30
	filled := barWidth * sc.Current / sc.Total
	unfilled := barWidth - filled

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	unfilledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	bar := filledStyle.Render(strings.Repeat("━", filled)) +
		unfilledStyle.Render(strings.Repeat("░", unfilled))

	label := fmt.Sprintf("  Step %d of %d", sc.Current, sc.Total)
	dimLabel := unfilledStyle.Render(label)

	r.last = fmt.Sprintf("\n%s%s\n", bar, dimLabel)
	return r.last, nil
}

func (r *ProgressRegion) Subscribe() []reflect.Type {
	return nil
}
