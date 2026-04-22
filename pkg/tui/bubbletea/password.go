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

type passwordModel struct {
	prompt      string
	textInput   textinput.Model
	submitted   bool
	aborted     bool
	interrupted bool // true for Ctrl+C (hard cancel), false for Esc (soft cancel)
}

func newPasswordModel(prompt string) passwordModel {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	return passwordModel{
		prompt:    prompt,
		textInput: ti,
	}
}

func (m passwordModel) Init() tea.Cmd { return textinput.Blink }

func (m passwordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case keyEnter:
			m.submitted = true
			return m, tea.Quit
		case keyCtrlC:
			m.interrupted = true
			return m, tea.Quit
		case keyEsc:
			m.aborted = true
			return m, func() tea.Msg { return GoBackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// Hints returns password-specific key hints for the hint bar.
func (m passwordModel) Hints() []string {
	return []string{"enter submit", "esc back"}
}

// Result returns the password value after the user submits.
func (m passwordModel) Result() (any, bool) {
	return m.textInput.Value(), m.submitted
}

// NewPasswordPrompt creates a password prompt model for use in the wizard composite.
func NewPasswordPrompt(prompt string) PromptModel {
	return newPasswordModel(prompt)
}

func (m passwordModel) View() tea.View {
	if m.submitted {
		return tea.NewView(fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), hintStyle.Render("[hidden]")))
	}
	return tea.NewView(fmt.Sprintf("%s %s\n%s", promptStyle.Render("?"), titleStyle.Render(m.prompt), m.textInput.View()))
}

// Password implements tui.Prompter.
func (p *Prompter) Password(ctx context.Context, prompt string) (string, error) {
	model := newPasswordModel(prompt)

	r := p.runProgram(ctx, model)
	if r.interrupted {
		return "", tui.ErrInterrupted
	}
	if r.err != nil {
		return "", fmt.Errorf("password prompt: %w", r.err)
	}

	m := r.model.(passwordModel)
	if m.aborted {
		return "", context.Canceled
	}
	return m.textInput.Value(), nil
}
