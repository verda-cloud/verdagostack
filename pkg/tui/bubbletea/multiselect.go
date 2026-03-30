package bubbletea

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

type multiSelectModel struct {
	prompt   string
	choices  []string
	cursor   int
	selected map[int]bool
	pageSize int
	loop     bool
	min      int
	max      int
	done     bool
	aborted  bool
	err      string
}

func newMultiSelectModel(prompt string, choices []string, cfg tui.MultiSelectConfig) multiSelectModel {
	ps := cfg.PageSize
	if ps <= 0 || ps > len(choices) {
		ps = len(choices)
	}
	selected := make(map[int]bool)
	for _, idx := range cfg.Defaults {
		if idx >= 0 && idx < len(choices) {
			selected[idx] = true
		}
	}
	return multiSelectModel{
		prompt:   prompt,
		choices:  choices,
		selected: selected,
		pageSize: ps,
		loop:     cfg.Loop,
		min:      cfg.Min,
		max:      cfg.Max,
	}
}

func (m multiSelectModel) Init() tea.Cmd { return nil }

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case " ":
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				if m.max > 0 && len(m.selected) >= m.max {
					m.err = fmt.Sprintf("maximum %d selections allowed", m.max)
				} else {
					m.selected[m.cursor] = true
					m.err = ""
				}
			}
		case "enter":
			if m.min > 0 && len(m.selected) < m.min {
				m.err = fmt.Sprintf("at least %d selections required", m.min)
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiSelectModel) View() string {
	if m.done {
		var names []string
		for i, c := range m.choices {
			if m.selected[i] {
				names = append(names, c)
			}
		}
		return fmt.Sprintf("? %s %s\n", m.prompt, strings.Join(names, ", "))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("? %s (space to toggle, enter to confirm)\n", m.prompt))

	start, end := m.visibleRange()
	for i := start; i < end; i++ {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}
		b.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, check, m.choices[i]))
	}

	if m.err != "" {
		b.WriteString(fmt.Sprintf("  ✗ %s\n", m.err))
	}
	return b.String()
}

func (m multiSelectModel) visibleRange() (int, int) {
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

// MultiSelect implements tui.Prompter.
func (p *Prompter) MultiSelect(ctx context.Context, prompt string, choices []string, opts ...tui.MultiSelectOption) ([]int, error) {
	if len(choices) == 0 {
		return nil, fmt.Errorf("multi-select prompt: no choices provided")
	}

	cfg := tui.ResolveMultiSelectConfig(opts)
	model := newMultiSelectModel(prompt, choices, cfg)

	program := tea.NewProgram(model,
		tea.WithInput(p.in),
		tea.WithOutput(p.out),
		tea.WithContext(ctx),
	)

	result, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("multi-select prompt: %w", err)
	}

	m := result.(multiSelectModel)
	if m.aborted {
		return nil, context.Canceled
	}

	var indices []int
	for i := range m.choices {
		if m.selected[i] {
			indices = append(indices, i)
		}
	}
	return indices, nil
}
