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

package bubbletea

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

// KeyBinding pairs a key matcher with its display label and handler.
// Generic in the model type so handlers can mutate model state directly.
//
// The label is a function of the current model state, enabling
// state-aware hint text (e.g. "esc clear filter" when a filter is
// active, "esc back" otherwise). An empty label hides the entry from
// the hint bar; the handler still runs.
type KeyBinding[M any] struct {
	// ID is a stable identifier for selective overrides via
	// configuration (relabel/hide/replace). Keep it short and kebab-case.
	ID string

	// Match returns true if the key event triggers this binding.
	// Pure function of the message — should not inspect model state.
	Match func(tea.KeyPressMsg) bool

	// Label returns the hint text for the current model state. Empty
	// string hides the entry from the hint bar. Pure function — no
	// side effects.
	Label func(*M) string

	// Handle runs the binding. Returns a tea.Cmd to issue (or nil)
	// and a "stop" flag — true means this Update tick is complete
	// and no further bindings should be tried. Return (nil, false)
	// to pass the event to the next matching binding.
	Handle func(*M, tea.KeyPressMsg) (tea.Cmd, bool)
}

// MatchKey returns a Match that fires when msg.Code equals any of the
// given key codes. Pass bubbletea's named constants (tea.KeyUp,
// tea.KeyEnter, …) or rune literals.
func MatchKey(codes ...rune) func(tea.KeyPressMsg) bool {
	return func(msg tea.KeyPressMsg) bool {
		return slices.Contains(codes, msg.Code)
	}
}

// MatchRune returns a Match for a printable rune optionally constrained
// by a modifier. MatchRune('c', tea.ModCtrl) matches Ctrl+C.
// With no modifier argument, the binding fires only when no modifier is
// held (so MatchRune('k') won't fire on Ctrl+K).
func MatchRune(r rune, mod ...tea.KeyMod) func(tea.KeyPressMsg) bool {
	var required tea.KeyMod
	for _, m := range mod {
		required |= m
	}
	return func(msg tea.KeyPressMsg) bool {
		if msg.Code != r {
			return false
		}
		if required == 0 {
			return msg.Mod == 0
		}
		return msg.Mod&required == required
	}
}

// MatchText returns a Match that fires on any key event carrying
// printable text (msg.Text != ""). Useful as a catch-all for
// type-to-filter style inputs. Order this binding after more specific
// rune matchers so they get first claim.
func MatchText() func(tea.KeyPressMsg) bool {
	return func(msg tea.KeyPressMsg) bool { return msg.Text != "" }
}

// Dispatch iterates bindings in order and invokes the first matching
// Handle that returns stop=true. Returns that handler's tea.Cmd and
// true; returns (nil, false) if no binding claims the event.
//
// A handler can choose to "pass through" by returning (nil, false)
// from its Handle even when its Match fires — useful when state
// matters (e.g. a vim 'k' binding that only navigates when the
// filter is empty).
func Dispatch[M any](m *M, bindings []KeyBinding[M], msg tea.KeyPressMsg) (tea.Cmd, bool) {
	for _, b := range bindings {
		if b.Match == nil || b.Handle == nil {
			continue
		}
		if !b.Match(msg) {
			continue
		}
		cmd, stop := b.Handle(m, msg)
		if stop {
			return cmd, true
		}
	}
	return nil, false
}

// HintsFor returns non-empty labels from the binding set for the given
// model state, preserving declaration order. Used to power Hints().
func HintsFor[M any](m *M, bindings []KeyBinding[M]) []string {
	out := make([]string, 0, len(bindings))
	for _, b := range bindings {
		if b.Label == nil {
			continue
		}
		if lbl := b.Label(m); lbl != "" {
			out = append(out, lbl)
		}
	}
	return out
}

// ApplyBindingOverrides returns a new binding slice with relabels
// applied (by ID) and hidden IDs filtered to empty-label closures.
// Defaults that don't appear in overrides pass through unchanged.
func ApplyBindingOverrides[M any](bindings []KeyBinding[M], relabels map[string]string, hidden []string) []KeyBinding[M] {
	if len(relabels) == 0 && len(hidden) == 0 {
		return bindings
	}
	hiddenSet := make(map[string]struct{}, len(hidden))
	for _, id := range hidden {
		hiddenSet[id] = struct{}{}
	}
	out := make([]KeyBinding[M], len(bindings))
	for i, b := range bindings {
		out[i] = b
		if _, ok := hiddenSet[b.ID]; ok {
			out[i].Label = func(*M) string { return "" }
			continue
		}
		if lbl, ok := relabels[b.ID]; ok {
			label := lbl
			out[i].Label = func(*M) string { return label }
		}
	}
	return out
}
