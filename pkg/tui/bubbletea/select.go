package bubbletea

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

type selectModel struct {
	prompt   string
	choices  []string
	cursor   int
	pageSize int
	loop     bool
	chosen   bool
	aborted  bool
}

func newSelectModel(prompt string, choices []string, cfg tui.SelectConfig) selectModel {
	ps := cfg.PageSize
	if ps <= 0 || ps > len(choices) {
		ps = len(choices)
	}
	cursor := cfg.Default
	if cursor < 0 || cursor >= len(choices) {
		cursor = 0
	}
	return selectModel{
		prompt:   prompt,
		choices:  choices,
		cursor:   cursor,
		pageSize: ps,
		loop:     cfg.Loop,
	}
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else if m.loop {
				m.cursor = len(m.choices) - 1
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			} else if m.loop {
				m.cursor = 0
			}
		case keyEnter:
			m.chosen = true
			return m, tea.Quit
		case keyCtrlC, keyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	if m.chosen {
		return fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), answerStyle.Render(m.choices[m.cursor]))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt))

	start, end := m.visibleRange()
	for i := start; i < end; i++ {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", cursorStyle.Render(">"), selectedStyle.Render(m.choices[i]))
		} else {
			fmt.Fprintf(&b, "    %s\n", dimStyle.Render(m.choices[i]))
		}
	}
	return b.String()
}

func (m selectModel) visibleRange() (int, int) {
	total := len(m.choices)
	if m.pageSize >= total {
		return 0, total
	}
	half := m.pageSize / 2
	start := m.cursor - half
	if start < 0 {
		start = 0
	}
	end := start + m.pageSize
	if end > total {
		end = total
		start = end - m.pageSize
	}
	return start, end
}

// Select implements tui.Prompter.
func (p *Prompter) Select(ctx context.Context, prompt string, choices []string, opts ...tui.SelectOption) (int, error) {
	if len(choices) == 0 {
		return -1, fmt.Errorf("select prompt: no choices provided")
	}

	cfg := tui.ResolveSelectConfig(opts)
	model := newSelectModel(prompt, choices, cfg)

	program := tea.NewProgram(model,
		tea.WithInput(p.in),
		tea.WithOutput(p.out),
		tea.WithContext(ctx),
	)

	result, err := program.Run()
	if err != nil {
		return -1, fmt.Errorf("select prompt: %w", err)
	}

	m := result.(selectModel)
	if m.aborted {
		return -1, context.Canceled
	}
	return m.cursor, nil
}
