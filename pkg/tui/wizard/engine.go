package wizard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

// Engine runs a wizard Flow interactively.
type Engine struct {
	prompter    tui.Prompter
	cache       *stepCache
	collected   map[string]any
	autoSkipped map[string]bool // steps auto-skipped due to empty loader results
	rewindCount map[string]int  // tracks rewinds per step to detect infinite loops
	current     int
	writer      io.Writer
}

// maxRewindsPerStep is the maximum number of times the engine will auto-rewind
// to the same step before giving up with an error.
const maxRewindsPerStep = 3

// NewEngine creates a wizard engine with the given prompter.
func NewEngine(prompter tui.Prompter, opts ...EngineOption) *Engine {
	e := &Engine{
		prompter:  prompter,
		cache:     newStepCache(),
		collected: make(map[string]any),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// EngineOption configures the Engine.
type EngineOption func(*Engine)

// WithOutput sets the writer for engine messages (defaults to os.Stderr).
func WithOutput(w io.Writer) EngineOption {
	return func(e *Engine) { e.writer = w }
}

func (e *Engine) out() io.Writer {
	if e.writer != nil {
		return e.writer
	}
	return os.Stderr
}

// Collected returns the values collected during the flow.
func (e *Engine) Collected() map[string]any {
	return e.collected
}

// Run executes the flow step by step.
func (e *Engine) Run(ctx context.Context, flow *Flow) error {
	e.current = 0
	e.collected = make(map[string]any)
	e.cache = newStepCache()
	e.autoSkipped = make(map[string]bool)
	e.rewindCount = make(map[string]int)
	for e.current < len(flow.Steps) {
		step := flow.Steps[e.current]

		// Check ShouldSkip BEFORE IsSet — a skipped step must be cleared
		// even if the user provided a value via flag/config.
		if step.ShouldSkip != nil && step.ShouldSkip(e.collected) {
			e.resetStep(step)
			e.current++
			continue
		}

		if step.IsSet != nil && step.IsSet() {
			// Propagate the pre-set value so downstream loaders/callbacks can see it.
			if step.Value != nil {
				val := step.Value()
				// Validate preset values — bad flags/config should not be silently accepted.
				if step.Validate != nil {
					if err := step.Validate(val); err != nil {
						return fmt.Errorf("step %q: preset value invalid: %w", step.Name, err)
					}
				}
				e.collected[step.Name] = val
			}
			e.current++
			continue
		}

		choices, err := e.loadStep(ctx, step)
		if err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}

		if step.Loader != nil && len(choices) == 0 {
			if !step.Required {
				// Optional step with empty choices — apply default or reset and skip.
				if step.Default != nil {
					val := step.Default(e.collected)
					if step.Setter != nil {
						step.Setter(val)
					}
					e.collected[step.Name] = val
				} else {
					e.resetStep(step)
				}
				e.autoSkipped[step.Name] = true
				e.current++
				continue
			}
			if e.current == 0 {
				return fmt.Errorf("step %q: no options available and cannot go back", step.Name)
			}
			e.rewindCount[step.Name]++
			if e.rewindCount[step.Name] > maxRewindsPerStep {
				return fmt.Errorf("step %q: no options available after %d attempts — the flow cannot proceed with current inputs", step.Name, maxRewindsPerStep)
			}
			_, _ = fmt.Fprintf(e.out(), "  No options available for %q — going back.\n", promptLabel(step))
			if err := e.goBackToDependency(flow, step); err != nil {
				return fmt.Errorf("step %q: %w", step.Name, err)
			}
			continue
		}

		canGoBack := e.hasEditablePriorStep(flow)
		value, err := e.prompt(ctx, step, choices, canGoBack)
		if errors.Is(err, errGoBack) {
			e.goBack(flow)
			continue
		}
		// Treat Esc/Ctrl+C (context.Canceled) as back-navigation when possible.
		if errors.Is(err, context.Canceled) {
			if canGoBack {
				e.goBack(flow)
				continue
			}
			return fmt.Errorf("wizard cancelled")
		}
		if err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}

		// For non-required steps, use default if value is empty.
		if !step.Required && isEmpty(value) && step.Default != nil {
			value = step.Default(e.collected)
		}

		// Enforce required: reject empty values.
		if step.Required && isEmpty(value) {
			_, _ = fmt.Fprintf(e.out(), "  %s is required — please provide a value.\n", promptLabel(step))
			continue // re-prompt same step
		}

		if step.Validate != nil {
			if err := step.Validate(value); err != nil {
				_, _ = fmt.Fprintf(e.out(), "  Validation error: %s\n", err)
				continue // re-prompt same step
			}
		}

		if step.Setter != nil {
			step.Setter(value)
		}
		e.collected[step.Name] = value
		delete(e.autoSkipped, step.Name)
		delete(e.rewindCount, step.Name)
		e.invalidateDownstream(flow, step.Name)
		e.current++
	}
	return nil
}

func (e *Engine) loadStep(ctx context.Context, step Step) ([]Choice, error) {
	if step.Loader == nil {
		return nil, nil
	}
	deps := e.depsFor(step)
	if cached, ok := e.cache.get(step.Name, deps); ok {
		return cached, nil
	}
	choices, err := step.Loader(ctx, e.prompter, e.collected)
	if err != nil {
		return nil, err
	}
	e.cache.set(step.Name, deps, choices)
	return choices, nil
}

// errGoBack is a sentinel indicating the user chose to go back.
var errGoBack = fmt.Errorf("go back")

// backLabel is appended to Select/MultiSelect choices when back navigation is possible.
const backLabel = "← Back"

// hasEditablePriorStep returns true if there is at least one earlier step
// that is not fixed (IsSet), not currently skipped (ShouldSkip), and not
// auto-skipped due to empty loader results.
func (e *Engine) hasEditablePriorStep(flow *Flow) bool {
	for i := e.current - 1; i >= 0; i-- {
		step := flow.Steps[i]
		if step.IsSet != nil && step.IsSet() {
			continue
		}
		if step.ShouldSkip != nil && step.ShouldSkip(e.collected) {
			continue
		}
		if e.autoSkipped[step.Name] {
			continue
		}
		return true
	}
	return false
}

func (e *Engine) prompt(ctx context.Context, step Step, choices []Choice, canGoBack bool) (any, error) {
	switch step.Prompt {
	case SelectPrompt:
		return e.promptSelect(ctx, step, choices, canGoBack)
	case MultiSelectPrompt:
		return e.promptMultiSelect(ctx, step, choices, canGoBack)
	case TextInputPrompt:
		return e.promptTextInput(ctx, step)
	case ConfirmPrompt:
		return e.promptConfirm(ctx, step)
	case PasswordPrompt:
		return e.prompter.Password(ctx, promptLabel(step))
	default:
		return nil, fmt.Errorf("unsupported prompt type: %d", step.Prompt)
	}
}

func (e *Engine) promptSelect(ctx context.Context, step Step, choices []Choice, canGoBack bool) (any, error) {
	labels := choiceLabels(choices)
	if canGoBack {
		labels = append(labels, backLabel)
	}
	var opts []tui.SelectOption
	if step.Default != nil {
		if defVal, ok := step.Default(e.collected).(string); ok {
			for i, c := range choices {
				if c.Value == defVal {
					opts = append(opts, tui.WithSelectDefault(i))
					break
				}
			}
		}
	}
	idx, err := e.prompter.Select(ctx, promptLabel(step), labels, opts...)
	if err != nil {
		return nil, err
	}
	if canGoBack && idx == len(choices) {
		return nil, errGoBack
	}
	return choices[idx].Value, nil
}

func (e *Engine) promptMultiSelect(ctx context.Context, step Step, choices []Choice, canGoBack bool) (any, error) {
	labels := choiceLabels(choices)
	if canGoBack {
		labels = append(labels, backLabel)
	}
	var opts []tui.MultiSelectOption
	if step.Required {
		opts = append(opts, tui.WithMinSelections(1))
	}
	if step.Default != nil {
		if defVals, ok := step.Default(e.collected).([]string); ok && len(defVals) > 0 {
			valSet := make(map[string]bool, len(defVals))
			for _, v := range defVals {
				valSet[v] = true
			}
			var defaults []int
			for i, c := range choices {
				if valSet[c.Value] {
					defaults = append(defaults, i)
				}
			}
			if len(defaults) > 0 {
				opts = append(opts, tui.WithMultiSelectDefaults(defaults))
			}
		}
	}
	indices, err := e.prompter.MultiSelect(ctx, promptLabel(step), labels, opts...)
	if err != nil {
		return nil, err
	}
	for _, idx := range indices {
		if canGoBack && idx == len(choices) {
			return nil, errGoBack
		}
	}
	values := make([]string, len(indices))
	for i, idx := range indices {
		values[i] = choices[idx].Value
	}
	return values, nil
}

func (e *Engine) promptTextInput(ctx context.Context, step Step) (any, error) {
	var opts []tui.TextInputOption
	if step.Default != nil {
		if d, ok := step.Default(e.collected).(string); ok && d != "" {
			opts = append(opts, tui.WithDefault(d))
		}
	}
	return e.prompter.TextInput(ctx, promptLabel(step), opts...)
}

func (e *Engine) promptConfirm(ctx context.Context, step Step) (any, error) {
	var opts []tui.ConfirmOption
	if step.Default != nil {
		if d, ok := step.Default(e.collected).(bool); ok {
			opts = append(opts, tui.WithConfirmDefault(d))
		}
	}
	return e.prompter.Confirm(ctx, promptLabel(step), opts...)
}

func (e *Engine) goBack(flow *Flow) {
	from := e.current
	e.current--
	for e.current >= 0 {
		step := flow.Steps[e.current]
		if step.IsSet != nil && step.IsSet() {
			e.current--
			continue
		}
		if step.ShouldSkip != nil && step.ShouldSkip(e.collected) {
			e.current--
			continue
		}
		if e.autoSkipped[step.Name] {
			e.current--
			continue
		}
		// Clear the destination step and all steps after it up to where we came from.
		for j := e.current; j < from; j++ {
			e.resetStep(flow.Steps[j])
		}
		return
	}
	e.current = 0
}

// goBackToDependency rewinds to the nearest editable step listed in the current
// step's DependsOn. Returns error only if ALL dependencies are fixed (IsSet).
func (e *Engine) goBackToDependency(flow *Flow, current Step) error {
	if len(current.DependsOn) == 0 {
		e.goBack(flow)
		return nil
	}

	depSet := make(map[string]bool, len(current.DependsOn))
	for _, d := range current.DependsOn {
		depSet[d] = true
	}

	// Scan all dependency steps, collect editable ones.
	// Skip deps that are fixed (IsSet) or currently hidden (ShouldSkip).
	nearestEditable := -1
	hasFixedDep := false
	hasSkippedDep := false
	for i := e.current - 1; i >= 0; i-- {
		step := flow.Steps[i]
		if !depSet[step.Name] {
			continue
		}
		if step.IsSet != nil && step.IsSet() {
			hasFixedDep = true
			continue // fixed via flag
		}
		if step.ShouldSkip != nil && step.ShouldSkip(e.collected) {
			hasSkippedDep = true
			continue // currently hidden — rewinding here would just skip again
		}
		if e.autoSkipped[step.Name] {
			hasSkippedDep = true
			continue // auto-skipped due to empty loader — not editable
		}
		if nearestEditable == -1 || i > nearestEditable {
			nearestEditable = i
		}
	}
	_ = hasSkippedDep // used in allFixed check below
	allFixed := hasFixedDep && !hasSkippedDep && nearestEditable < 0

	if nearestEditable >= 0 {
		// Clear everything from the editable dep up to current.
		for j := nearestEditable; j < e.current; j++ {
			e.resetStep(flow.Steps[j])
		}
		e.current = nearestEditable
		return nil
	}

	// No editable dependency found. If all are truly fixed (IsSet), error out.
	if allFixed {
		return fmt.Errorf("step %q has no options available and all its dependencies %v are fixed (set via flag)", current.Name, current.DependsOn)
	}

	// Dependencies are currently skipped. Find the earliest editable step
	// that could change the skip condition.
	if hasSkippedDep {
		for i := 0; i < e.current; i++ {
			step := flow.Steps[i]
			if step.IsSet != nil && step.IsSet() {
				continue
			}
			if step.ShouldSkip != nil && step.ShouldSkip(e.collected) {
				continue
			}
			if e.autoSkipped[step.Name] {
				continue
			}
			// Found the earliest editable step — rewind to it.
			for j := i; j < e.current; j++ {
				e.resetStep(flow.Steps[j])
			}
			e.current = i
			return nil
		}
		// All steps before current are fixed or skipped — no rewind target.
		return fmt.Errorf("step %q has no options available and no editable prior step exists to change the outcome", current.Name)
	}

	// No dependencies matched at all — should not happen, but fail safely.
	return fmt.Errorf("step %q has no options available and cannot go back", current.Name)
}

func (e *Engine) invalidateDownstream(flow *Flow, changedStep string) {
	for _, step := range flow.Steps {
		for _, dep := range step.DependsOn {
			if dep == changedStep {
				e.cache.invalidate(step.Name)
				break
			}
		}
	}
}

func (e *Engine) depsFor(step Step) map[string]any {
	deps := make(map[string]any, len(step.DependsOn))
	for _, name := range step.DependsOn {
		if v, ok := e.collected[name]; ok {
			deps[name] = v
		}
	}
	return deps
}

// resetStep clears a step's value from collected and calls its Resetter.
func (e *Engine) resetStep(step Step) {
	delete(e.collected, step.Name)
	if step.Resetter != nil {
		step.Resetter()
	}
}

// promptLabel returns the display text for a step, falling back to Name if Description is empty.
func promptLabel(step Step) string {
	if step.Description != "" {
		return step.Description
	}
	return step.Name
}

func choiceLabels(choices []Choice) []string {
	labels := make([]string, len(choices))
	for i, c := range choices {
		labels[i] = c.Label
	}
	return labels
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []string:
		return len(val) == 0
	default:
		return false
	}
}
