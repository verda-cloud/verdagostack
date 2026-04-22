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

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

type confirmModel struct {
	prompt      string
	value       bool
	decided     bool
	aborted     bool
	interrupted bool // true for Ctrl+C (hard cancel), false for Esc (soft cancel)
}

func newConfirmModel(prompt string, cfg tui.ConfirmConfig) confirmModel {
	return confirmModel{
		prompt: prompt,
		value:  cfg.Default,
	}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y":
			m.value = true
			m.decided = true
			return m, tea.Quit
		case "n", "N":
			m.value = false
			m.decided = true
			return m, tea.Quit
		case keyEnter:
			m.decided = true
			return m, tea.Quit
		case keyCtrlC:
			m.interrupted = true
			return m, tea.Quit
		case keyEsc:
			m.aborted = true
			return m, func() tea.Msg { return GoBackMsg{} }
		}
	}
	return m, nil
}

// Hints returns confirm-specific key hints for the hint bar.
func (m confirmModel) Hints() []string {
	return []string{"y/n", "enter confirm", "esc back"}
}

// Result returns the confirm value after the user decides.
func (m confirmModel) Result() (any, bool) {
	return m.value, m.decided
}

// NewConfirmPrompt creates a confirm prompt model for use in the wizard composite.
func NewConfirmPrompt(prompt string, cfg tui.ConfirmConfig) PromptModel {
	return newConfirmModel(prompt, cfg)
}

func (m confirmModel) View() tea.View {
	hint := "y/N"
	if m.value {
		hint = "Y/n"
	}
	if m.decided || m.aborted {
		answer := "No"
		if m.value {
			answer = "Yes"
		}
		return tea.NewView(fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), answerStyle.Render(answer)))
	}
	return tea.NewView(fmt.Sprintf("%s %s %s ", promptStyle.Render("?"), titleStyle.Render(m.prompt), hintStyle.Render("["+hint+"]")))
}

// Confirm implements tui.Prompter.
func (p *Prompter) Confirm(ctx context.Context, prompt string, opts ...tui.ConfirmOption) (bool, error) {
	cfg := tui.ResolveConfirmConfig(opts)
	model := newConfirmModel(prompt, cfg)

	r := p.runProgram(ctx, model)
	if r.interrupted {
		return false, tui.ErrInterrupted
	}
	if r.err != nil {
		return false, fmt.Errorf("confirm prompt: %w", r.err)
	}

	m := r.model.(confirmModel)
	if m.aborted {
		return false, context.Canceled
	}
	return m.value, nil
}
