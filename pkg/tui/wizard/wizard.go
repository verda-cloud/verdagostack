package wizard

import (
	"context"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

// PromptType defines which TUI widget to use for a step.
type PromptType int

const (
	SelectPrompt      PromptType = iota // Single selection from a list.
	MultiSelectPrompt                   // Multiple selections from a list.
	TextInputPrompt                     // Free-form text input.
	ConfirmPrompt                       // Yes/No confirmation.
	PasswordPrompt                      // Masked text input.
)

// Choice represents one selectable option in a list prompt.
type Choice struct {
	Label       string // Displayed to the user.
	Value       string // Actual value stored in collected map.
	Description string // Optional extra info.
}

// LoaderFunc fetches available choices for a step.
// It receives the Prompter so it can run sub-prompts internally
// (e.g., SSH key creation when no keys exist).
// The collected map contains values from all previously completed steps.
type LoaderFunc func(ctx context.Context, prompter tui.Prompter, collected map[string]any) ([]Choice, error)

// Step defines one step in the wizard flow.
type Step struct {
	Name        string
	Description string
	Prompt      PromptType
	Required    bool
	Default     func(collected map[string]any) any
	ShouldSkip  func(collected map[string]any) bool
	Loader      LoaderFunc
	Validate    func(value any) error
	Setter      func(value any)
	Resetter    func()      // Called when step value is cleared (back/skip). Resets the bound variable.
	IsSet       func() bool // Returns true if value was provided via flag/config.
	Value       func() any  // Returns the current value when IsSet is true. Propagates to collected map.
	DependsOn   []string
}

// Flow defines a complete wizard execution graph.
type Flow struct {
	Name  string
	Steps []Step
}

// StaticChoices returns a LoaderFunc that always returns the given choices.
// Use for steps with fixed options that don't require an API call.
func StaticChoices(choices ...Choice) LoaderFunc {
	return func(_ context.Context, _ tui.Prompter, _ map[string]any) ([]Choice, error) {
		return choices, nil
	}
}
