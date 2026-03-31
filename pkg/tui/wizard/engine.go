package wizard

import (
	"context"
	"fmt"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

// Engine runs a wizard Flow interactively.
type Engine struct {
	prompter  tui.Prompter
	cache     *stepCache
	collected map[string]any
	current   int
}

// NewEngine creates a wizard engine with the given prompter.
func NewEngine(prompter tui.Prompter) *Engine {
	return &Engine{
		prompter:  prompter,
		cache:     newStepCache(),
		collected: make(map[string]any),
	}
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
			e.current++
			continue
		}

		choices, err := e.loadStep(ctx, step)
		if err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}

		if step.Required && step.Loader != nil && len(choices) == 0 {
			if e.current == 0 {
				return fmt.Errorf("step %q: no options available and cannot go back", step.Name)
			}
			e.goBack(flow)
			continue
		}

		value, err := e.prompt(ctx, step, choices)
		if err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}

		if !step.Required && isEmpty(value) && step.Default != nil {
			value = step.Default(e.collected)
		}

		if step.Validate != nil {
			if err := step.Validate(value); err != nil {
				return fmt.Errorf("step %q validation: %w", step.Name, err)
			}
		}

		step.Setter(value)
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

func (e *Engine) prompt(ctx context.Context, step Step, choices []Choice) (any, error) {
	switch step.Prompt {
	case SelectPrompt:
		labels := choiceLabels(choices)
		idx, err := e.prompter.Select(ctx, step.Description, labels)
		if err != nil {
			return nil, err
		}
		return choices[idx].Value, nil

	case MultiSelectPrompt:
		labels := choiceLabels(choices)
		indices, err := e.prompter.MultiSelect(ctx, step.Description, labels)
		if err != nil {
			return nil, err
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
		delete(e.collected, step.Name)
		return
	}
	e.current = 0
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
