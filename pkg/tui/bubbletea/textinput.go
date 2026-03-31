package bubbletea

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

type textInputModel struct {
	prompt    string
	textInput textinput.Model
	submitted bool
	aborted   bool
	validate  func(string) error
	err       error
}

func newTextInputModel(prompt string, cfg tui.TextInputConfig) textInputModel {
	ti := textinput.New()
	ti.Placeholder = cfg.Placeholder
	ti.SetValue(cfg.Default)
	ti.Focus()

	return textInputModel{
		prompt:    prompt,
		textInput: ti,
		validate:  cfg.Validate,
	}
}

func (m textInputModel) Init() tea.Cmd { return textinput.Blink }

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		case keyCtrlC, keyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.err = nil
	return m, cmd
}

func (m textInputModel) View() string {
	if m.submitted {
		return fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), answerStyle.Render(m.textInput.Value()))
	}
	s := fmt.Sprintf("%s %s\n%s", promptStyle.Render("?"), titleStyle.Render(m.prompt), m.textInput.View())
	if m.err != nil {
		s += fmt.Sprintf("\n  %s", errorStyle.Render("✗ "+m.err.Error()))
	}
	return s
}

// TextInput implements tui.Prompter.
func (p *Prompter) TextInput(ctx context.Context, prompt string, opts ...tui.TextInputOption) (string, error) {
	cfg := tui.ResolveTextInputConfig(opts)
	model := newTextInputModel(prompt, cfg)

	program := tea.NewProgram(model,
		tea.WithInput(p.in),
		tea.WithOutput(p.out),
		tea.WithContext(ctx),
	)

	result, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("text input prompt: %w", err)
	}

	m := result.(textInputModel)
	if m.aborted {
		return "", context.Canceled
	}
	return m.textInput.Value(), nil
}
