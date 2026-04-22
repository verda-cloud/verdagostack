// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tui

import (
	"context"
	"io"
)

// Prompter is the core interface for interactive terminal prompts.
// Implementations must be safe for sequential use but need not be concurrent-safe.
type Prompter interface {
	// Confirm asks a yes/no question and returns the boolean answer.
	Confirm(ctx context.Context, prompt string, opts ...ConfirmOption) (bool, error)

	// TextInput asks the user for a single line of text.
	TextInput(ctx context.Context, prompt string, opts ...TextInputOption) (string, error)

	// Password asks the user for sensitive input (masked).
	Password(ctx context.Context, prompt string) (string, error)

	// Select presents a list of choices and returns the selected index.
	Select(ctx context.Context, prompt string, choices []string, opts ...SelectOption) (int, error)

	// MultiSelect presents a list of choices and returns selected indices.
	MultiSelect(ctx context.Context, prompt string, choices []string, opts ...MultiSelectOption) ([]int, error)

	// Editor opens a multi-line text editor and returns the result.
	Editor(ctx context.Context, prompt string, opts ...EditorOption) (string, error)
}

// IO holds the input/output streams for the prompter.
// This mirrors the pattern used in pkg/app for testability.
type IO struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// DefaultIO returns IO wired to os.Stdin/Stdout/Stderr.
func DefaultIO() IO {
	return IO{
		In:     nil, // nil means os.Stdin at runtime — set by implementation
		Out:    nil,
		ErrOut: nil,
	}
}

// --- Confirm Options ---

// ConfirmOption configures a Confirm prompt.
type ConfirmOption func(*ConfirmConfig)

// ConfirmConfig holds resolved Confirm settings.
type ConfirmConfig struct {
	Default bool // default answer when user presses Enter
}

// WithConfirmDefault sets the default value for a confirm prompt.
func WithConfirmDefault(v bool) ConfirmOption {
	return func(c *ConfirmConfig) { c.Default = v }
}

// --- TextInput Options ---

// TextInputOption configures a TextInput prompt.
type TextInputOption func(*TextInputConfig)

// TextInputConfig holds resolved TextInput settings.
type TextInputConfig struct {
	Default     string
	Placeholder string
	Validate    func(string) error
}

// WithDefault sets the default value for a text input.
func WithDefault(v string) TextInputOption {
	return func(c *TextInputConfig) { c.Default = v }
}

// WithPlaceholder sets the placeholder text.
func WithPlaceholder(v string) TextInputOption {
	return func(c *TextInputConfig) { c.Placeholder = v }
}

// WithValidation sets a validation function for the input.
func WithValidation(fn func(string) error) TextInputOption {
	return func(c *TextInputConfig) { c.Validate = fn }
}

// --- Select Options ---

// SelectOption configures a Select prompt.
type SelectOption func(*SelectConfig)

// SelectConfig holds resolved Select settings.
type SelectConfig struct {
	Default  int  // index of the default selection
	PageSize int  // number of visible items (0 = show all)
	Loop     bool // wrap around at ends
}

// WithSelectDefault sets the default selected index.
func WithSelectDefault(index int) SelectOption {
	return func(c *SelectConfig) { c.Default = index }
}

// WithPageSize sets the number of visible items.
func WithPageSize(n int) SelectOption {
	return func(c *SelectConfig) { c.PageSize = n }
}

// WithLoop enables wrapping when navigating past ends.
func WithLoop(v bool) SelectOption {
	return func(c *SelectConfig) { c.Loop = v }
}

// --- MultiSelect Options ---

// MultiSelectOption configures a MultiSelect prompt.
type MultiSelectOption func(*MultiSelectConfig)

// MultiSelectConfig holds resolved MultiSelect settings.
type MultiSelectConfig struct {
	Defaults []int // indices selected by default
	PageSize int
	Loop     bool
	Min      int // minimum required selections (0 = no minimum)
	Max      int // maximum allowed selections (0 = no maximum)
}

// WithMultiSelectDefaults sets the default selected indices.
func WithMultiSelectDefaults(indices []int) MultiSelectOption {
	return func(c *MultiSelectConfig) { c.Defaults = indices }
}

// WithMultiSelectPageSize sets the number of visible items.
func WithMultiSelectPageSize(n int) MultiSelectOption {
	return func(c *MultiSelectConfig) { c.PageSize = n }
}

// WithMinSelections sets the minimum required selections.
func WithMinSelections(n int) MultiSelectOption {
	return func(c *MultiSelectConfig) { c.Min = n }
}

// WithMaxSelections sets the maximum allowed selections.
func WithMaxSelections(n int) MultiSelectOption {
	return func(c *MultiSelectConfig) { c.Max = n }
}

// --- Editor Options ---

// EditorOption configures an Editor prompt.
type EditorOption func(*EditorConfig)

// EditorConfig holds resolved Editor settings.
type EditorConfig struct {
	Default  string
	FileExt  string // file extension hint for syntax highlighting
	ShowHelp bool
}

// WithEditorDefault sets the initial content in the editor.
func WithEditorDefault(v string) EditorOption {
	return func(c *EditorConfig) { c.Default = v }
}

// WithFileExt sets the file extension hint.
func WithFileExt(ext string) EditorOption {
	return func(c *EditorConfig) { c.FileExt = ext }
}
