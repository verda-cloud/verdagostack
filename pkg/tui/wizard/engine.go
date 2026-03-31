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
	prompter  tui.Prompter
	cache     *stepCache
	collected map[string]any
	current   int
	writer    io.Writer
}

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
	for e.current < len(flow.Steps) {
		step := flow.Steps[e.current]

		if step.IsSet != nil && step.IsSet() {
			e.current++
			continue
		}

		if step.ShouldSkip != nil && step.ShouldSkip(e.collected) {
			// Clear any stale value from a previous pass through this step.
			if _, hadValue := e.collected[step.Name]; hadValue {
				e.resetStep(step)
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
				// Optional step with empty choices — apply default and skip.
				if step.Default != nil {
					val := step.Default(e.collected)
					if step.Setter != nil {
						step.Setter(val)
					}
					e.collected[step.Name] = val
				}
				e.current++
				continue
			}
			if e.current == 0 {
				return fmt.Errorf("step %q: no options available and cannot go back", step.Name)
			}
			_, _ = fmt.Fprintf(e.out(), "  No options available for %q — going back.\n", step.Description)
			if err := e.goBackToDependency(flow, step); err != nil {
				return fmt.Errorf("step %q: %w", step.Name, err)
			}
			continue
		}

		value, err := e.prompt(ctx, step, choices)
		if errors.Is(err, errGoBack) {
			e.goBack(flow)
			continue
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
			_, _ = fmt.Fprintf(e.out(), "  %s is required — please provide a value.\n", step.Description)
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

func (e *Engine) prompt(ctx context.Context, step Step, choices []Choice) (any, error) {
	canGoBack := e.current > 0

	switch step.Prompt {
	case SelectPrompt:
		labels := choiceLabels(choices)
		if canGoBack {
			labels = append(labels, backLabel)
		}
		idx, err := e.prompter.Select(ctx, step.Description, labels)
		if err != nil {
			return nil, err
		}
		if canGoBack && idx == len(choices) {
			return nil, errGoBack
		}
		return choices[idx].Value, nil

	case MultiSelectPrompt:
		labels := choiceLabels(choices)
		if canGoBack {
			labels = append(labels, backLabel)
		}
		var msOpts []tui.MultiSelectOption
		if step.Required {
			msOpts = append(msOpts, tui.WithMinSelections(1))
		}
		indices, err := e.prompter.MultiSelect(ctx, step.Description, labels, msOpts...)
		if err != nil {
			return nil, err
		}
		// If "← Back" is among selected indices, go back.
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

	case TextInputPrompt:
		var opts []tui.TextInputOption
		if step.Default != nil {
			if d, ok := step.Default(e.collected).(string); ok && d != "" {
				opts = append(opts, tui.WithDefault(d))
			}
		}
		return e.prompter.TextInput(ctx, step.Description, opts...)

	case ConfirmPrompt:
		return e.prompter.Confirm(ctx, step.Description)

	case PasswordPrompt:
		return e.prompter.Password(ctx, step.Description)

	default:
		return nil, fmt.Errorf("unsupported prompt type: %d", step.Prompt)
	}
}

func (e *Engine) goBack(flow *Flow) {
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
		e.resetStep(step)
		return
	}
	e.current = 0
}

// goBackToDependency rewinds to the nearest step listed in the current step's
// DependsOn. This skips over intervening text/confirm steps so the user lands
// on the step that actually controls the empty result.
// Returns error if all dependencies are fixed (IsSet) and cannot be changed.
func (e *Engine) goBackToDependency(flow *Flow, current Step) error {
	if len(current.DependsOn) == 0 {
		e.goBack(flow)
		return nil
	}

	depSet := make(map[string]bool, len(current.DependsOn))
	for _, d := range current.DependsOn {
		depSet[d] = true
	}

	// Scan backwards for an editable dependency.
	for i := e.current - 1; i >= 0; i-- {
		step := flow.Steps[i]
		if !depSet[step.Name] {
			continue
		}
		// Found a dependency — check if it's editable.
		if step.IsSet != nil && step.IsSet() {
			// This dependency is fixed via flag — can't change it.
			// Return error so the wizard stops instead of looping.
			return fmt.Errorf("step %q has no options available and its dependency %q is fixed (set via flag)", current.Name, step.Name)
		}
		// Editable dependency found — clear it and everything after up to current.
		for j := i; j < e.current; j++ {
			e.resetStep(flow.Steps[j])
		}
		e.current = i
		return nil
	}

	// No dependency found, fall back to regular goBack.
	e.goBack(flow)
	return nil
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
