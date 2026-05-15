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

// LiveLister is an optional capability interface for prompters that
// support a live-updating picker. Kept separate from Prompter to avoid
// breaking implementations (test fakes, alternate engines) that don't
// support live updates. Callers should type-assert:
//
//	if ll, ok := prompter.(tui.LiveLister); ok {
//	    idx, err := ll.LiveList(ctx, prompt, rows, updates, opts...)
//	} else {
//	    // fall back to Prompter.Select with pre-fetched labels
//	}
//
// Semantics:
//   - Updates with an unknown Key are silently dropped.
//   - Multiple updates for the same Key apply in arrival order
//     (last-write-wins).
//   - Cursor identity is preserved across updates: hovering Key
//     "foo" stays on "foo" even if its filtered index shifts.
//   - Type-to-filter matches against the current (possibly updated)
//     Label, so rows can drop in/out of the filtered view.
//   - The caller owns concurrency, retry, and error handling. The
//     renderer styles rows where Err != nil distinctly but still
//     shows the supplied Label.
//   - Closing the channel is a courtesy — the prompter never depends
//     on close to function correctly.
type LiveLister interface {
	LiveList(ctx context.Context, prompt string, rows []LiveRow, updates <-chan LiveListUpdate, opts ...LiveListOption) (int, error)
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
	Default       bool              // default answer when user presses Enter
	RelabelByID   map[string]string // override individual binding labels by ID
	HiddenByID    []string          // suppress these binding labels from the hint bar
	ExtraBindings any               // engine-specific extra/replacement bindings
}

// WithConfirmDefault sets the default value for a confirm prompt.
func WithConfirmDefault(v bool) ConfirmOption {
	return func(c *ConfirmConfig) { c.Default = v }
}

// WithConfirmRelabel renames a single binding's hint by its stable ID
// (e.g. "yes-no", "confirm", "esc", "exit").
func WithConfirmRelabel(id, label string) ConfirmOption {
	return func(c *ConfirmConfig) {
		if c.RelabelByID == nil {
			c.RelabelByID = make(map[string]string)
		}
		c.RelabelByID[id] = label
	}
}

// WithConfirmHide suppresses one or more bindings from the hint bar by ID.
func WithConfirmHide(ids ...string) ConfirmOption {
	return func(c *ConfirmConfig) {
		c.HiddenByID = append(c.HiddenByID, ids...)
	}
}

// --- TextInput Options ---

// TextInputOption configures a TextInput prompt.
type TextInputOption func(*TextInputConfig)

// TextInputConfig holds resolved TextInput settings.
type TextInputConfig struct {
	Default       string
	Placeholder   string
	Validate      func(string) error
	RelabelByID   map[string]string // override individual binding labels by ID
	HiddenByID    []string          // suppress these binding labels from the hint bar
	ExtraBindings any               // engine-specific extra/replacement bindings
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

// WithTextInputRelabel renames a single binding's hint by its stable ID
// (e.g. "submit", "esc", "exit").
func WithTextInputRelabel(id, label string) TextInputOption {
	return func(c *TextInputConfig) {
		if c.RelabelByID == nil {
			c.RelabelByID = make(map[string]string)
		}
		c.RelabelByID[id] = label
	}
}

// WithTextInputHide suppresses one or more bindings from the hint bar by ID.
func WithTextInputHide(ids ...string) TextInputOption {
	return func(c *TextInputConfig) {
		c.HiddenByID = append(c.HiddenByID, ids...)
	}
}

// --- Select Options ---

// SelectOption configures a Select prompt.
type SelectOption func(*SelectConfig)

// SelectConfig holds resolved Select settings.
type SelectConfig struct {
	Default       int               // index of the default selection
	PageSize      int               // number of visible items (0 = show all)
	Loop          bool              // wrap around at ends
	ShowHints     bool              // render the prompt's Hints() bar below the choices
	Hints         []string          // override hint strings entirely; nil = use prompt defaults
	RelabelByID   map[string]string // override individual binding labels by ID
	HiddenByID    []string          // suppress these binding labels from the hint bar
	ExtraBindings any               // engine-specific extra/replacement bindings (set via bubbletea options)
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

// WithShowHints toggles rendering the prompt's built-in key hints below the
// choices. Off by default — wizard flows render hints externally and must
// not set this. Use for non-wizard interactive prompts (list browsers,
// pickers) where the wizard composite isn't in play.
func WithShowHints(v bool) SelectOption {
	return func(c *SelectConfig) { c.ShowHints = v }
}

// WithHints overrides the hint strings rendered by the hint bar (when
// ShowHints is on). Useful for localization or shorter labels. Key handling
// is unaffected — this only changes what's displayed.
//
// WithHints replaces ALL hints. For per-binding rename, use WithSelectRelabel.
// For suppression only, use WithSelectHide.
func WithHints(hints ...string) SelectOption {
	return func(c *SelectConfig) { c.Hints = hints }
}

// WithSelectRelabel renames a single binding's hint by its stable ID
// (e.g. "navigate", "select", "esc", "exit"). Use this for localization
// without rewriting the entire hint set. Unknown IDs are ignored.
func WithSelectRelabel(id, label string) SelectOption {
	return func(c *SelectConfig) {
		if c.RelabelByID == nil {
			c.RelabelByID = make(map[string]string)
		}
		c.RelabelByID[id] = label
	}
}

// WithSelectHide suppresses one or more bindings from the hint bar by ID.
// The key still works — only its label is hidden.
func WithSelectHide(ids ...string) SelectOption {
	return func(c *SelectConfig) {
		c.HiddenByID = append(c.HiddenByID, ids...)
	}
}

// --- LiveList Options ---

// LiveRow is a row in a LiveList. Key identifies the row stably across
// label updates. Label is the initial rendered string and what
// type-to-filter matches against until an update replaces it.
type LiveRow struct {
	Key   string
	Label string
}

// LiveListUpdate replaces the Label for the row identified by Key.
// When Err is non-nil the renderer styles the row body distinctly;
// callers should still provide a meaningful Label in the error case
// (e.g. "error: rate limited") since Err is signal, not text.
type LiveListUpdate struct {
	Key   string
	Label string
	Err   error
}

// LiveListOption configures a LiveList prompt.
type LiveListOption func(*LiveListConfig)

// LiveListConfig mirrors SelectConfig for keyboard/hint behavior.
// Kept as a separate type so adding LiveList-only fields (debounce
// interval, etc.) doesn't widen SelectConfig.
type LiveListConfig struct {
	Default       int               // initial cursor index (post-filter)
	PageSize      int               // visible rows (0 = show all)
	Loop          bool              // wrap cursor at ends
	ShowHints     bool              // render Hints() bar below choices
	Hints         []string          // override hints entirely; nil = defaults from prompt
	RelabelByID   map[string]string // rename specific bindings by ID
	HiddenByID    []string          // suppress specific bindings from the hint bar
	ExtraBindings any               // engine-specific extras (set via bubbletea options)
}

// WithLiveListDefault sets the initial cursor position.
func WithLiveListDefault(index int) LiveListOption {
	return func(c *LiveListConfig) { c.Default = index }
}

// WithLiveListPageSize sets the number of visible rows.
func WithLiveListPageSize(n int) LiveListOption {
	return func(c *LiveListConfig) { c.PageSize = n }
}

// WithLiveListLoop enables cursor wrap.
func WithLiveListLoop(v bool) LiveListOption {
	return func(c *LiveListConfig) { c.Loop = v }
}

// WithLiveListShowHints toggles rendering the prompt's built-in key
// hints below the rows. Off by default — wizard flows render hints
// externally and must not set this.
func WithLiveListShowHints(v bool) LiveListOption {
	return func(c *LiveListConfig) { c.ShowHints = v }
}

// WithLiveListHints overrides the hint strings entirely (when ShowHints
// is on). Replaces ALL hints. Use WithLiveListRelabel for per-binding
// rename instead.
func WithLiveListHints(hints ...string) LiveListOption {
	return func(c *LiveListConfig) { c.Hints = hints }
}

// WithLiveListRelabel renames a single binding's hint by its stable ID.
// Same ID space as Select (navigate, select, esc, exit, …).
func WithLiveListRelabel(id, label string) LiveListOption {
	return func(c *LiveListConfig) {
		if c.RelabelByID == nil {
			c.RelabelByID = make(map[string]string)
		}
		c.RelabelByID[id] = label
	}
}

// WithLiveListHide suppresses one or more bindings from the hint bar
// by ID. The key still works — only its label is hidden.
func WithLiveListHide(ids ...string) LiveListOption {
	return func(c *LiveListConfig) {
		c.HiddenByID = append(c.HiddenByID, ids...)
	}
}

// --- MultiSelect Options ---

// MultiSelectOption configures a MultiSelect prompt.
type MultiSelectOption func(*MultiSelectConfig)

// MultiSelectConfig holds resolved MultiSelect settings.
type MultiSelectConfig struct {
	Defaults      []int // indices selected by default
	PageSize      int
	Loop          bool
	Min           int               // minimum required selections (0 = no minimum)
	Max           int               // maximum allowed selections (0 = no maximum)
	ShowHints     bool              // render the prompt's Hints() bar below the choices
	Hints         []string          // override hint strings entirely; nil = use prompt defaults
	RelabelByID   map[string]string // override individual binding labels by ID
	HiddenByID    []string          // suppress these binding labels from the hint bar
	ExtraBindings any               // engine-specific extra/replacement bindings (set via bubbletea options)
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

// WithMultiSelectShowHints toggles rendering the prompt's built-in key hints
// below the choices. Off by default — wizard flows render hints externally
// and must not set this. Use for non-wizard interactive prompts where the
// wizard composite isn't in play.
func WithMultiSelectShowHints(v bool) MultiSelectOption {
	return func(c *MultiSelectConfig) { c.ShowHints = v }
}

// WithMultiSelectHints overrides the hint strings rendered by the hint bar
// (when ShowHints is on). Useful for localization or shorter labels. Key
// handling is unaffected — this only changes what's displayed.
//
// WithMultiSelectHints replaces ALL hints. For per-binding rename, use
// WithMultiSelectRelabel. For suppression only, use WithMultiSelectHide.
func WithMultiSelectHints(hints ...string) MultiSelectOption {
	return func(c *MultiSelectConfig) { c.Hints = hints }
}

// WithMultiSelectRelabel renames a single binding's hint by its stable ID
// (e.g. "navigate", "toggle", "select-all", "confirm", "esc", "exit").
// Use for localization without rewriting the full hint set.
func WithMultiSelectRelabel(id, label string) MultiSelectOption {
	return func(c *MultiSelectConfig) {
		if c.RelabelByID == nil {
			c.RelabelByID = make(map[string]string)
		}
		c.RelabelByID[id] = label
	}
}

// WithMultiSelectHide suppresses one or more bindings from the hint bar
// by ID. The key still works — only its label is hidden.
func WithMultiSelectHide(ids ...string) MultiSelectOption {
	return func(c *MultiSelectConfig) {
		c.HiddenByID = append(c.HiddenByID, ids...)
	}
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
