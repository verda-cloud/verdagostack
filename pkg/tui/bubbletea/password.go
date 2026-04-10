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
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
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
