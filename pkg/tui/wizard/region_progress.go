package wizard

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// ProgressRegionOption configures a ProgressRegion.
type ProgressRegionOption func(*ProgressRegion)

// WithProgressGradient sets the gradient colors for the progress bar.
// Defaults to the bubbles default gradient (#5A56E0 → #EE6FF8).
func WithProgressGradient(colorA, colorB string) ProgressRegionOption {
	return func(r *ProgressRegion) {
		r.colorA = colorA
		r.colorB = colorB
	}
}

// WithProgressSolidFill uses a single color instead of a gradient.
func WithProgressSolidFill(color string) ProgressRegionOption {
	return func(r *ProgressRegion) {
		r.solidFill = color
	}
}

// WithProgressWidth sets the bar width in characters (default: 40).
func WithProgressWidth(w int) ProgressRegionOption {
	return func(r *ProgressRegion) {
		r.width = w
	}
}

// WithProgressPercent shows percentage text (e.g., "33%") instead of the
// default "Step X of Y" label. Follows the bubbletea animated progress example.
func WithProgressPercent() ProgressRegionOption {
	return func(r *ProgressRegion) {
		r.showPercent = true
		r.hideStepLabel = true
	}
}

// WithoutProgressStepLabel hides the "Step X of Y" label.
func WithoutProgressStepLabel() ProgressRegionOption {
	return func(r *ProgressRegion) {
		r.hideStepLabel = true
	}
}

// ProgressRegion displays an animated-style step progress bar using
// the charmbracelet/bubbles progress component for gradient rendering.
// Responds to StepChangedMsg.
type ProgressRegion struct {
	last          string
	colorA        string
	colorB        string
	solidFill     string
	width         int
	showPercent   bool
	hideStepLabel bool
}

// NewProgressRegion creates a progress bar region.
func NewProgressRegion(opts ...ProgressRegionOption) *ProgressRegion {
	r := &ProgressRegion{
		width: 40,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *ProgressRegion) buildBar() progress.Model {
	var opts []progress.Option
	opts = append(opts, progress.WithWidth(r.width))
	if !r.showPercent {
		opts = append(opts, progress.WithoutPercentage())
	}
	if r.solidFill != "" {
		opts = append(opts, progress.WithSolidFill(r.solidFill))
	} else if r.colorA != "" && r.colorB != "" {
		opts = append(opts, progress.WithGradient(r.colorA, r.colorB))
	} else {
		opts = append(opts, progress.WithDefaultGradient())
	}
	return progress.New(opts...)
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

	pct := float64(sc.Current) / float64(sc.Total)
	bar := r.buildBar()

	rendered := bar.ViewAs(pct)
	if !r.hideStepLabel {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		label := fmt.Sprintf("  Step %d of %d", sc.Current, sc.Total)
		rendered += dimStyle.Render(label)
	}

	r.last = fmt.Sprintf("\n%s\n", rendered)
	return r.last, nil
}

func (r *ProgressRegion) Subscribe() []reflect.Type {
	return nil
}
