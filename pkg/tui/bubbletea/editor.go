package bubbletea

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

type editorModel struct {
	prompt    string
	textarea  textarea.Model
	submitted bool
	aborted   bool
}

func newEditorModel(prompt string, cfg tui.EditorConfig) editorModel {
	ta := textarea.New()
	ta.SetValue(cfg.Default)
	ta.ShowLineNumbers = true
	ta.Focus()

	return editorModel{
		prompt:   prompt,
		textarea: ta,
	}
}

func (m editorModel) Init() tea.Cmd { return textarea.Blink }

func (m editorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			m.submitted = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m editorModel) View() string {
	if m.submitted {
		lines := strings.Count(m.textarea.Value(), "\n") + 1
		return fmt.Sprintf("? %s [%d lines]\n", m.prompt, lines)
	}
	return fmt.Sprintf("? %s (ctrl+d to submit, esc to cancel)\n%s", m.prompt, m.textarea.View())
}

// Editor implements tui.Prompter.
func (p *Prompter) Editor(ctx context.Context, prompt string, opts ...tui.EditorOption) (string, error) {
	cfg := tui.ResolveEditorConfig(opts)
	model := newEditorModel(prompt, cfg)

	program := tea.NewProgram(model,
		tea.WithInput(p.in),
		tea.WithOutput(p.out),
		tea.WithContext(ctx),
	)

	result, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("editor prompt: %w", err)
	}

	m := result.(editorModel)
	if m.aborted {
		return "", context.Canceled
	}
	return m.textarea.Value(), nil
}
