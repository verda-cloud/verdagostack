package bubbletea

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type passwordModel struct {
	prompt    string
	textInput textinput.Model
	submitted bool
	aborted   bool
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
	case tea.KeyMsg:
		switch msg.String() {
		case keyEnter:
			m.submitted = true
			return m, tea.Quit
		case keyCtrlC, keyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m passwordModel) View() string {
	if m.submitted {
		return fmt.Sprintf("? %s [hidden]\n", m.prompt)
	}
	return fmt.Sprintf("? %s\n%s", m.prompt, m.textInput.View())
}

// Password implements tui.Prompter.
func (p *Prompter) Password(ctx context.Context, prompt string) (string, error) {
	model := newPasswordModel(prompt)

	program := tea.NewProgram(model,
		tea.WithInput(p.in),
		tea.WithOutput(p.out),
		tea.WithContext(ctx),
	)

	result, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("password prompt: %w", err)
	}

	m := result.(passwordModel)
	if m.aborted {
		return "", context.Canceled
	}
	return m.textInput.Value(), nil
}
