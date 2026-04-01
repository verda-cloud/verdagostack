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
	choices  []string // original full list — never mutated
	filter   string   // current filter text
	matched  []int    // indices into choices that match filter
	cursor   int      // position within matched (not choices)
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
	matched := make([]int, len(choices))
	for i := range choices {
		matched[i] = i
	}
	return selectModel{
		prompt:   prompt,
		choices:  choices,
		filter:   "",
		matched:  matched,
		cursor:   cursor,
		pageSize: ps,
		loop:     cfg.Loop,
	}
}

func (m *selectModel) refilter() {
	if m.filter == "" {
		m.matched = make([]int, len(m.choices))
		for i := range m.choices {
			m.matched[i] = i
		}
	} else {
		lower := strings.ToLower(m.filter)
		m.matched = m.matched[:0]
		for i, c := range m.choices {
			if strings.Contains(strings.ToLower(c), lower) {
				m.matched = append(m.matched, i)
			}
		}
	}
	m.cursor = 0
}

func (m *selectModel) moveUp() {
	if len(m.matched) == 0 {
		return
	}
	if m.cursor > 0 {
		m.cursor--
	} else if m.loop {
		m.cursor = len(m.matched) - 1
	}
}

func (m *selectModel) moveDown() {
	if len(m.matched) == 0 {
		return
	}
	if m.cursor < len(m.matched)-1 {
		m.cursor++
	} else if m.loop {
		m.cursor = 0
	}
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.moveUp()
		case tea.KeyDown:
			m.moveDown()
		case tea.KeyEnter:
			if len(m.matched) == 0 {
				return m, nil
			}
			m.chosen = true
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.aborted = true
			return m, tea.Quit
		case tea.KeyEscape:
			if m.filter != "" {
				m.filter = ""
				m.refilter()
			} else {
				m.aborted = true
				return m, tea.Quit
			}
		case tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.refilter()
			}
		case tea.KeyRunes:
			s := string(msg.Runes)
			if m.filter == "" {
				switch s {
				case "k":
					m.moveUp()
					return m, nil
				case "j":
					m.moveDown()
					return m, nil
				}
			}
			m.filter += s
			m.refilter()
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	if m.chosen {
		selected := m.choices[m.matched[m.cursor]]
		return fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), answerStyle.Render(selected))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", promptStyle.Render("?"), titleStyle.Render(m.prompt))
	if m.filter != "" {
		fmt.Fprintf(&b, " %s", answerStyle.Render(m.filter))
	}
	b.WriteString("\n")

	if len(m.matched) == 0 {
		fmt.Fprintf(&b, "    %s\n", dimStyle.Render("no matches"))
		return b.String()
	}

	start, end := m.visibleRange()
	for i := start; i < end; i++ {
		label := m.choices[m.matched[i]]
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", cursorStyle.Render(">"), selectedStyle.Render(label))
		} else {
			fmt.Fprintf(&b, "    %s\n", dimStyle.Render(label))
		}
	}
	return b.String()
}

func (m selectModel) visibleRange() (int, int) {
	total := len(m.matched)
	if total == 0 {
		return 0, 0
	}
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
	return m.matched[m.cursor], nil
}
