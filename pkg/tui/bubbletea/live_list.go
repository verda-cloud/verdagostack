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
	"context"
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

// liveListUpdateMsg is the bubbletea-message form of LiveListUpdate.
// Routed onto the program from the pump goroutine in Prompter.LiveList.
type liveListUpdateMsg struct {
	Key   string
	Label string
	Err   error
}

// liveListModel composes selectModel for view/filter/cursor state and
// adds a Key→index map plus a per-row error set. It overrides Update,
// View, and Hints so dispatch and rendering use its own binding set
// rather than the embedded selectModel's defaults.
type liveListModel struct {
	selectModel
	keyIndex map[string]int              // Key → index in choices/keys
	keys     []string                    // Key at each index (parallel to choices)
	errs     map[int]bool                // index → errored
	bindings []KeyBinding[liveListModel] // shadows selectModel.bindings; this slice drives dispatch + Hints
}

// DefaultLiveListBindings returns the canonical binding set: identical
// to DefaultSelectBindings, re-typed for *liveListModel via the
// embedded selectModel. Exported so callers can reuse/reorder.
func DefaultLiveListBindings() []KeyBinding[liveListModel] {
	return adaptSelectBindings(DefaultSelectBindings())
}

// adaptSelectBindings re-types a []KeyBinding[selectModel] as
// []KeyBinding[liveListModel] by routing Label/Handle through the
// embedded selectModel. Internal helper.
func adaptSelectBindings(src []KeyBinding[selectModel]) []KeyBinding[liveListModel] {
	out := make([]KeyBinding[liveListModel], len(src))
	for i, b := range src {
		out[i] = KeyBinding[liveListModel]{
			ID:    b.ID,
			Match: b.Match,
			Label: func(m *liveListModel) string {
				if b.Label == nil {
					return ""
				}
				return b.Label(&m.selectModel)
			},
			Handle: func(m *liveListModel, msg tea.KeyPressMsg) (tea.Cmd, bool) {
				if b.Handle == nil {
					return nil, false
				}
				return b.Handle(&m.selectModel, msg)
			},
		}
	}
	return out
}

// WithLiveListAddBindings prepends extra bindings, giving them priority
// over the defaults. Mirrors WithSelectAddBindings.
func WithLiveListAddBindings(extras ...KeyBinding[liveListModel]) tui.LiveListOption {
	return func(c *tui.LiveListConfig) {
		existing, _ := c.ExtraBindings.([]KeyBinding[liveListModel])
		c.ExtraBindings = append(existing, extras...)
	}
}

func newLiveListModel(prompt string, rows []tui.LiveRow, cfg tui.LiveListConfig) liveListModel {
	choices := make([]string, len(rows))
	keys := make([]string, len(rows))
	keyIndex := make(map[string]int, len(rows))
	for i, r := range rows {
		choices[i] = r.Label
		keys[i] = r.Key
		keyIndex[r.Key] = i
	}
	// Construct the embedded selectModel with the navigation-relevant
	// fields; binding-level overrides are applied at the liveListModel
	// layer below so liveListModel.bindings is the active set.
	sm := newSelectModel(prompt, choices, tui.SelectConfig{
		Default:   cfg.Default,
		PageSize:  cfg.PageSize,
		Loop:      cfg.Loop,
		ShowHints: cfg.ShowHints,
		Hints:     cfg.Hints,
	})
	defaults := ApplyBindingOverrides(DefaultLiveListBindings(), cfg.RelabelByID, cfg.HiddenByID)
	var bindings []KeyBinding[liveListModel]
	if extras, ok := cfg.ExtraBindings.([]KeyBinding[liveListModel]); ok && len(extras) > 0 {
		bindings = make([]KeyBinding[liveListModel], 0, len(extras)+len(defaults))
		bindings = append(bindings, extras...)
		bindings = append(bindings, defaults...)
	} else {
		bindings = defaults
	}
	return liveListModel{
		selectModel: sm,
		keyIndex:    keyIndex,
		keys:        keys,
		errs:        make(map[int]bool),
		bindings:    bindings,
	}
}

func (m liveListModel) Init() tea.Cmd { return nil }

func (m liveListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case GoBackMsg:
		// Standalone mode quit; wizard composite intercepts before this.
		return m, tea.Quit
	case liveListUpdateMsg:
		idx, ok := m.keyIndex[v.Key]
		if !ok {
			return m, nil // unknown key — drop
		}
		m.choices[idx] = v.Label
		if v.Err != nil {
			m.errs[idx] = true
		} else {
			delete(m.errs, idx)
		}
		// Preserve cursor by Key: snapshot key under cursor, refilter,
		// restore cursor to that key's new position in the matched set.
		cursorKey := ""
		if len(m.matched) > 0 && m.cursor >= 0 && m.cursor < len(m.matched) {
			cursorKey = m.keys[m.matched[m.cursor]]
		}
		m.refilter()
		if cursorKey != "" {
			for i, mi := range m.matched {
				if m.keys[mi] == cursorKey {
					m.cursor = i
					break
				}
			}
		}
		return m, nil
	case tea.KeyPressMsg:
		cmd, _ := Dispatch(&m, m.bindings, v)
		return m, cmd
	}
	return m, nil
}

func (m liveListModel) View() tea.View {
	if m.chosen {
		selected := m.choices[m.matched[m.cursor]]
		return tea.NewView(fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), answerStyle.Render(selected)))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", promptStyle.Render("?"), titleStyle.Render(m.prompt))
	if m.filter != "" {
		fmt.Fprintf(&b, " %s", answerStyle.Render(m.filter))
	}
	b.WriteString("\n")

	if len(m.matched) == 0 {
		fmt.Fprintf(&b, "    %s\n", dimStyle.Render("no matches"))
		m.renderLiveHintBar(&b)
		return tea.NewView(b.String())
	}

	start, end := m.visibleRange()
	for i := start; i < end; i++ {
		choiceIdx := m.matched[i]
		label := m.choices[choiceIdx]
		errored := m.errs[choiceIdx]
		switch {
		case i == m.cursor && errored:
			fmt.Fprintf(&b, "  %s %s\n", cursorStyle.Render(">"), errorStyle.Render(label))
		case i == m.cursor:
			fmt.Fprintf(&b, "  %s %s\n", cursorStyle.Render(">"), selectedStyle.Render(label))
		case errored:
			fmt.Fprintf(&b, "    %s\n", errorStyle.Render(label))
		default:
			fmt.Fprintf(&b, "    %s\n", dimStyle.Render(label))
		}
	}
	m.renderLiveHintBar(&b)
	return tea.NewView(b.String())
}

// renderLiveHintBar mirrors selectModel.renderHintBar but uses
// liveListModel.Hints() so the live-list's own binding overrides
// drive the displayed text.
func (m liveListModel) renderLiveHintBar(b *strings.Builder) {
	if !m.showHints {
		return
	}
	fmt.Fprintf(b, "\n%s\n", dimStyle.Render(strings.Join(m.Hints(), " · ")))
}

// Hints returns hints derived from liveListModel.bindings (not the
// embedded selectModel's). A non-nil customHints (via WithLiveListHints)
// still wins.
func (m liveListModel) Hints() []string {
	if m.customHints != nil {
		return m.customHints
	}
	return HintsFor(&m, m.bindings)
}

// Result returns the selected index after the user presses Enter.
// Inherited semantics from selectModel — index is into the original
// rows slice (not the post-filter matched set).
func (m liveListModel) Result() (any, bool) {
	if !m.chosen || len(m.matched) == 0 {
		return nil, false
	}
	return m.matched[m.cursor], true
}

// NewLiveListPrompt creates a live-list prompt model for use in the
// wizard composite. Updates must be delivered as liveListUpdateMsg
// values via the composite's runtime (not commonly used yet).
func NewLiveListPrompt(prompt string, rows []tui.LiveRow, cfg tui.LiveListConfig) PromptModel {
	return newLiveListModel(prompt, rows, cfg)
}

// pumpLiveListUpdates forwards LiveListUpdate values from updates to
// send until any of the exit conditions fires:
//   - ctx is canceled
//   - done is closed (the bubbletea program has returned)
//   - the updates channel is closed
//
// Extracted from LiveList so each exit path can be unit-tested in
// isolation (driving a real bubbletea program from a test is awkward
// without a TTY).
func pumpLiveListUpdates(ctx context.Context, done <-chan struct{}, updates <-chan tui.LiveListUpdate, send func(liveListUpdateMsg)) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case u, ok := <-updates:
			if !ok {
				return
			}
			send(liveListUpdateMsg{Key: u.Key, Label: u.Label, Err: u.Err})
		}
	}
}

// LiveList implements tui.Prompter.
func (p *Prompter) LiveList(ctx context.Context, prompt string, rows []tui.LiveRow, updates <-chan tui.LiveListUpdate, opts ...tui.LiveListOption) (int, error) {
	if len(rows) == 0 {
		return -1, fmt.Errorf("live list prompt: no rows provided")
	}

	cfg := tui.ResolveLiveListConfig(opts)
	model := newLiveListModel(prompt, rows, cfg)

	prog := tea.NewProgram(model,
		tea.WithInput(p.in),
		tea.WithOutput(p.out),
		tea.WithContext(ctx),
	)

	// Pump updates onto the program. Exits on any of: ctx done, channel
	// closed, or the program returning (the done channel closed below).
	// The done signal prevents the goroutine from lingering past
	// prog.Run when the caller leaves both ctx alive and updates unclosed.
	done := make(chan struct{})
	if updates != nil {
		go pumpLiveListUpdates(ctx, done, updates, func(m liveListUpdateMsg) { prog.Send(m) })
	}

	result, err := prog.Run()
	close(done)
	interrupted := errors.Is(err, tea.ErrInterrupted)
	if !interrupted {
		if m, ok := result.(liveListModel); ok {
			interrupted = m.interrupted
		}
	}
	if interrupted {
		return -1, tui.ErrInterrupted
	}
	if err != nil {
		return -1, fmt.Errorf("live list prompt: %w", err)
	}

	m := result.(liveListModel)
	if m.aborted {
		return -1, context.Canceled
	}
	if !m.chosen || len(m.matched) == 0 {
		return -1, context.Canceled
	}
	return m.matched[m.cursor], nil
}
