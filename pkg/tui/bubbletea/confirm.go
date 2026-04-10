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
			return m, tea.Quit
		}
	}
	return m, nil
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
