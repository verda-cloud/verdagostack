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
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

type textInputModel struct {
	prompt      string
	textInput   textinput.Model
	submitted   bool
	aborted     bool
	interrupted bool // true for Ctrl+C (hard cancel), false for Esc (soft cancel)
	pristine    bool // true until user types first character; typing clears the pre-filled default
	validate    func(string) error
	err         error
}

func newTextInputModel(prompt string, cfg tui.TextInputConfig) textInputModel {
	ti := textinput.New()
	ti.Placeholder = cfg.Placeholder
	ti.SetValue(cfg.Default)
	ti.Focus()

	return textInputModel{
		prompt:    prompt,
		textInput: ti,
		pristine:  cfg.Default != "",
		validate:  cfg.Validate,
	}
}

func (m textInputModel) Init() tea.Cmd { return textinput.Blink }

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case keyEnter:
			if m.validate != nil {
				if err := m.validate(m.textInput.Value()); err != nil {
					m.err = err
					return m, nil
				}
			}
			m.submitted = true
			return m, tea.Quit
		case keyCtrlC:
			m.interrupted = true
			return m, tea.Quit
		case keyEsc:
			m.aborted = true
			return m, func() tea.Msg { return GoBackMsg{} }
		default:
			// Clear pre-filled default on first keystroke so the user's
			// input replaces it rather than appending. This mimics the
			// native "select-all on focus" behavior.
			if m.pristine {
				m.pristine = false
				if msg.Text != "" && msg.Mod == 0 {
					m.textInput.SetValue("")
				}
			}
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.err = nil
	return m, cmd
}

// Hints returns text input-specific key hints for the hint bar.
func (m textInputModel) Hints() []string {
	return []string{"enter submit", "esc back"}
}

// Result returns the text input value after the user submits.
func (m textInputModel) Result() (any, bool) {
	return m.textInput.Value(), m.submitted
}

// NewTextInputPrompt creates a text input prompt model for use in the wizard composite.
func NewTextInputPrompt(prompt string, cfg tui.TextInputConfig) PromptModel {
	return newTextInputModel(prompt, cfg)
}

func (m textInputModel) View() tea.View {
	if m.submitted {
		return tea.NewView(fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), answerStyle.Render(m.textInput.Value())))
	}
	s := fmt.Sprintf("%s %s\n%s", promptStyle.Render("?"), titleStyle.Render(m.prompt), m.textInput.View())
	if m.err != nil {
		s += fmt.Sprintf("\n  %s", errorStyle.Render("✗ "+m.err.Error()))
	}
	return tea.NewView(s)
}

// TextInput implements tui.Prompter.
func (p *Prompter) TextInput(ctx context.Context, prompt string, opts ...tui.TextInputOption) (string, error) {
	cfg := tui.ResolveTextInputConfig(opts)
	model := newTextInputModel(prompt, cfg)

	r := p.runProgram(ctx, model)
	if r.interrupted {
		return "", tui.ErrInterrupted
	}
	if r.err != nil {
		return "", fmt.Errorf("text input prompt: %w", r.err)
	}

	m := r.model.(textInputModel)
	if m.aborted {
		return "", context.Canceled
	}
	return m.textInput.Value(), nil
}
