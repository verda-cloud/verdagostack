package wizard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

// stepState represents the lifecycle state of a step during execution.
type stepState int

const (
	statePending     stepState = iota // not yet visited
	stateFixed                        // IsSet=true, value from flag/config
	stateSkipped                      // ShouldSkip returned true
	stateAutoSkipped                  // loader returned empty choices (optional step)
	stateCompleted                    // user answered or default applied
)

// stepRuntime holds per-step execution state.
type stepRuntime struct {
	state       stepState
	value       any      // the collected value
	choices     []Choice // cached loader results
	loaded      bool     // true if loader has been called (distinguishes nil from not-loaded)
	rewindCount int      // how many times auto-rewind has targeted this step
}

// maxRewindsPerStep is the maximum number of times the engine will auto-rewind
// to the same step before giving up with an error.
const maxRewindsPerStep = 3

// Engine runs a wizard Flow interactively.
type Engine struct {
	prompter tui.Prompter
	status   tui.Status
	store    *Store
	bus      *MessageBus
	flow     *Flow         // the flow being executed (set during Run)
	steps    []stepRuntime // per-step runtime state, indexed same as flow.Steps
	current  int
	writer   io.Writer
}

// EngineOption configures the Engine.
type EngineOption func(*Engine)

// WithOutput sets the writer for engine messages (defaults to os.Stderr).
func WithOutput(w io.Writer) EngineOption {
	return func(e *Engine) { e.writer = w }
}

// NewEngine creates a wizard engine with the given prompter and optional status.
func NewEngine(prompter tui.Prompter, status tui.Status, opts ...EngineOption) *Engine {
	e := &Engine{
		prompter: prompter,
		status:   status,
		store:    NewStore(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Store returns the engine's shared store.
func (e *Engine) Store() *Store {
	return e.store
}

func (e *Engine) out() io.Writer {
	if e.writer != nil {
		return e.writer
	}
	return os.Stderr
}

// Collected returns the values collected during the flow.
func (e *Engine) Collected() map[string]any {
	return e.store.Collected()
}

// Run executes the flow step by step.
func (e *Engine) Run(ctx context.Context, flow *Flow) error {
	e.current = 0
	e.flow = flow
	e.steps = make([]stepRuntime, len(flow.Steps))
	e.store.Reset()

	// Initialize message bus with layout regions (or default).
	e.bus = NewMessageBus()
	layout := flow.Layout
	if layout == nil {
		layout = []RegionDef{{ID: "progress", Region: NewProgressRegion()}}
	}
	for _, def := range layout {
		e.bus.Register(def.ID, def.Region)
	}

	// Wire store change notifications to the message bus.
	e.store.onChange = func(key string, value any) {
		e.bus.Broadcast(StoreChangedMsg{Key: key, Value: value})
	}

	for e.current < len(flow.Steps) {
		step := flow.Steps[e.current]
		col := e.store.Collected()

		// ShouldSkip takes priority over IsSet.
		if step.ShouldSkip != nil && step.ShouldSkip(col) {
			e.transition(e.current, stateSkipped, nil)
			e.current++
			continue
		}

		if step.IsSet != nil && step.IsSet() {
			if err := e.handleFixed(step); err != nil {
				return err
			}
			e.current++
			continue
		}

		choices, err := e.loadChoices(ctx, step)
		if err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}

		if step.Loader != nil && len(choices) == 0 {
			handled, err := e.handleEmptyChoices(step)
			if err != nil {
				return err
			}
			if handled {
				continue
			}
		}

		if err := e.handlePrompt(ctx, step, choices); err != nil {
			return err
		}
	}
	return nil
}

// handleFixed processes a step with IsSet=true.
func (e *Engine) handleFixed(step Step) error {
	if step.Value != nil {
		val := step.Value()
		if step.Validate != nil {
			if err := step.Validate(val); err != nil {
				return fmt.Errorf("step %q: preset value invalid: %w", step.Name, err)
			}
		}
		if step.Setter != nil {
			step.Setter(val)
		}
		e.steps[e.current].state = stateFixed
		e.steps[e.current].value = val
		e.store.SetCollected(step.Name, val)
	} else {
		e.steps[e.current].state = stateFixed
	}
	return nil
}

// handleEmptyChoices handles the case when a step's loader returns no choices.
func (e *Engine) handleEmptyChoices(step Step) (bool, error) {
	col := e.store.Collected()
	if !step.Required {
		if step.Default != nil {
			val := step.Default(col)
			if step.Setter != nil {
				step.Setter(val)
			}
			e.transition(e.current, stateAutoSkipped, val)
		} else {
			e.transition(e.current, stateAutoSkipped, nil)
		}
		e.current++
		return true, nil
	}

	if e.current == 0 {
		return false, fmt.Errorf("step %q: no options available and cannot go back", step.Name)
	}

	e.steps[e.current].rewindCount++
	if e.steps[e.current].rewindCount > maxRewindsPerStep {
		return false, fmt.Errorf("step %q: no options available after %d attempts — the flow cannot proceed with current inputs", step.Name, maxRewindsPerStep)
	}

	_, _ = fmt.Fprintf(e.out(), "  No options available for %q — going back.\n", promptLabel(step))
	if err := e.rewindToDependency(step); err != nil {
		return false, fmt.Errorf("step %q: %w", step.Name, err)
	}
	return true, nil
}

// handlePrompt prompts the user and processes the result.
func (e *Engine) handlePrompt(ctx context.Context, step Step, choices []Choice) error {
	// Broadcast step change to all regions and render.
	// Use absolute position (e.current+1 of len(Steps)) so the bar is
	// stable — it always increments and the total never changes, even
	// when steps are skipped or fixed.
	e.bus.Broadcast(StepChangedMsg{
		Current:   e.current + 1,
		Total:     len(e.flow.Steps),
		StepName:  step.Name,
		Collected: e.store.Collected(),
	})
	e.renderRegions()

	canGoBack := e.hasEditablePriorStep()
	value, err := e.prompt(ctx, step, choices, canGoBack)

	if errors.Is(err, errGoBack) {
		e.rewindOne()
		return nil
	}
	if errors.Is(err, context.Canceled) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if canGoBack {
			e.rewindOne()
			return nil
		}
		return fmt.Errorf("wizard cancelled")
	}
	if err != nil {
		return fmt.Errorf("step %q: %w", step.Name, err)
	}

	col := e.store.Collected()

	// Apply default for non-required empty values.
	if !step.Required && isEmpty(value) && step.Default != nil {
		value = step.Default(col)
	}

	// Enforce required.
	if step.Required && isEmpty(value) {
		_, _ = fmt.Fprintf(e.out(), "  %s is required — please provide a value.\n", promptLabel(step))
		return nil // re-prompt same step (current not advanced)
	}

	// Validate.
	if step.Validate != nil {
		if err := step.Validate(value); err != nil {
			_, _ = fmt.Fprintf(e.out(), "  Validation error: %s\n", err)
			return nil // re-prompt
		}
	}

	// Complete the step.
	if step.Setter != nil {
		step.Setter(value)
	}
	e.transition(e.current, stateCompleted, value)

	// Broadcast collected change to all regions and re-render.
	e.bus.Broadcast(CollectedChangedMsg{
		Key:       step.Name,
		Value:     value,
		Collected: e.store.Collected(),
	})
	e.renderRegions()

	e.invalidateDownstream(e.current)
	e.current++
	return nil
}

// --- State transitions ---

// transition sets a step to a new state, resetting the caller-bound variable if needed.
func (e *Engine) transition(idx int, newState stepState, value any) {
	rt := &e.steps[idx]
	step := e.flow.Steps[idx]

	// Call Resetter when clearing a step's value (moving to pending, skipped,
	// or auto-skipped without a value). This ensures the caller-bound variable
	// is always consistent with the engine state.
	shouldReset := (newState == statePending || newState == stateSkipped) ||
		(newState == stateAutoSkipped && value == nil)
	if shouldReset && step.Resetter != nil {
		step.Resetter()
	}

	rt.state = newState
	rt.value = value
	rt.choices = nil // invalidate cached choices
	rt.loaded = false
	rt.rewindCount = 0 // reset guard so revisits get fresh attempts

	// Keep store in sync.
	if value != nil {
		e.store.SetCollected(step.Name, value)
	} else {
		e.store.ClearCollected(step.Name)
	}
}

// resetRange resets all non-fixed steps in [from, to) to statePending.
func (e *Engine) resetRange(from, to int) {
	for i := from; i < to; i++ {
		if e.steps[i].state == stateFixed {
			continue // preserve preset values
		}
		e.transition(i, statePending, nil)
	}
}

// --- Navigation ---

// rewindOne goes back to the nearest editable prior step, clearing everything between.
func (e *Engine) rewindOne() {
	from := e.current
	e.current--
	for e.current >= 0 {
		if e.isEditable(e.current) {
			e.resetRange(e.current, from+1) // +1 to include the abandoned step
			return
		}
		e.current--
	}
	e.current = 0
}

// rewindToDependency goes to the nearest editable dependency, or the earliest
// editable step if deps are skipped. Returns error if no rewind target exists.
func (e *Engine) rewindToDependency(current Step) error {
	if len(current.DependsOn) == 0 {
		e.rewindOne()
		return nil
	}

	depSet := make(map[string]bool, len(current.DependsOn))
	for _, d := range current.DependsOn {
		depSet[d] = true
	}

	// Classify dependencies.
	nearestEditable := -1
	hasFixed := false
	hasSkipped := false
	for i := e.current - 1; i >= 0; i-- {
		step := e.flow.Steps[i]
		if !depSet[step.Name] {
			continue
		}
		switch {
		case e.steps[i].state == stateFixed:
			hasFixed = true
		case !e.isEditable(i):
			hasSkipped = true
		default:
			if nearestEditable == -1 || i > nearestEditable {
				nearestEditable = i
			}
		}
	}

	// Direct editable dependency found.
	if nearestEditable >= 0 {
		e.resetRange(nearestEditable, e.current)
		e.current = nearestEditable
		return nil
	}

	// All deps are fixed — unrecoverable.
	if hasFixed && !hasSkipped {
		return fmt.Errorf("step %q has no options available and all its dependencies %v are fixed (set via flag)", current.Name, current.DependsOn)
	}

	// Deps are skipped — find the earliest editable step that could change
	// the skip condition. We use earliest (not nearest) because the skip
	// condition typically depends on an earlier choice like "environment".
	if hasSkipped {
		for i := 0; i < e.current; i++ {
			if e.isEditable(i) {
				e.resetRange(i, e.current+1)
				e.current = i
				return nil
			}
		}
		return fmt.Errorf("step %q has no options available and no editable prior step exists to change the outcome", current.Name)
	}

	// Fallback.
	e.rewindOne()
	return nil
}

// isEditable returns true if a step can be prompted (not fixed, not skipped, not auto-skipped).
func (e *Engine) isEditable(idx int) bool {
	step := e.flow.Steps[idx]
	rt := e.steps[idx]
	if step.IsSet != nil && step.IsSet() {
		return false
	}
	col := e.store.Collected()
	if step.ShouldSkip != nil && step.ShouldSkip(col) {
		return false
	}
	if rt.state == stateAutoSkipped {
		return false
	}
	return true
}

// hasEditablePriorStep returns true if there is at least one earlier editable step.
func (e *Engine) hasEditablePriorStep() bool {
	for i := e.current - 1; i >= 0; i-- {
		if e.isEditable(i) {
			return true
		}
	}
	return false
}

// --- Choice loading ---

func (e *Engine) loadChoices(ctx context.Context, step Step) ([]Choice, error) {
	if step.Loader == nil {
		return nil, nil
	}

	// Check cache — loaded distinguishes "not yet called" from "called, returned nil".
	rt := &e.steps[e.current]
	if rt.loaded {
		return rt.choices, nil
	}

	choices, err := step.Loader(ctx, e.prompter, e.status, e.store)
	if err != nil {
		return nil, err
	}
	rt.choices = choices
	rt.loaded = true
	return choices, nil
}

func (e *Engine) invalidateDownstream(changedIdx int) {
	changedName := e.flow.Steps[changedIdx].Name
	for i, step := range e.flow.Steps {
		for _, dep := range step.DependsOn {
			if dep == changedName {
				e.steps[i].choices = nil
				e.steps[i].loaded = false
				break
			}
		}
	}
}

// --- Prompting ---

// errGoBack is a sentinel indicating the user chose to go back.
var errGoBack = fmt.Errorf("go back")

// backLabel is appended to Select/MultiSelect choices when back navigation is possible.
const backLabel = "← Back"

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
		col := e.store.Collected() // safe: only needs completed values
		if defVal, ok := step.Default(col).(string); ok {
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
		col := e.store.Collected()
		if defVals, ok := step.Default(col).([]string); ok && len(defVals) > 0 {
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
		col := e.store.Collected()
		if d, ok := step.Default(col).(string); ok && d != "" {
			opts = append(opts, tui.WithDefault(d))
		}
	}
	return e.prompter.TextInput(ctx, promptLabel(step), opts...)
}

func (e *Engine) promptConfirm(ctx context.Context, step Step) (any, error) {
	var opts []tui.ConfirmOption
	if step.Default != nil {
		col := e.store.Collected()
		if d, ok := step.Default(col).(bool); ok {
			opts = append(opts, tui.WithConfirmDefault(d))
		}
	}
	return e.prompter.Confirm(ctx, promptLabel(step), opts...)
}

// --- Rendering ---

// renderRegions outputs region renders that changed since the last call.
func (e *Engine) renderRegions() {
	for _, output := range e.bus.RenderChanged() {
		if output != "" {
			_, _ = fmt.Fprint(e.out(), output)
		}
	}
}

// --- Utilities ---

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
