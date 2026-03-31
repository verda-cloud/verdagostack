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
	IsSet       func() bool
	DependsOn   []string
}

// Flow defines a complete wizard execution graph.
type Flow struct {
	Name  string
	Steps []Step
}
