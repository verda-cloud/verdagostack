package bubbletea

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

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
	case tea.KeyPressMsg:
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
		case " ", "space":
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
		case keyEnter:
			if m.min > 0 && len(m.selected) < m.min {
				m.err = fmt.Sprintf("at least %d selections required", m.min)
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case keyCtrlC, keyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiSelectModel) View() tea.View {
	if m.done {
		var names []string
		for i, c := range m.choices {
			if m.selected[i] {
				names = append(names, c)
			}
		}
		return tea.NewView(fmt.Sprintf("%s %s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt), answerStyle.Render(strings.Join(names, ", "))))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", promptStyle.Render("?"), titleStyle.Render(m.prompt))

	start, end := m.visibleRange()
	for i := start; i < end; i++ {
		cur := "  "
		if i == m.cursor {
			cur = cursorStyle.Render("> ")
		}
		check := uncheckStyle.Render("[ ]")
		label := dimStyle.Render(m.choices[i])
		if m.selected[i] {
			check = checkStyle.Render("[x]")
			label = selectedStyle.Render(m.choices[i])
		}
		if i == m.cursor && !m.selected[i] {
			label = selectedStyle.Render(m.choices[i])
		}
		fmt.Fprintf(&b, "  %s%s %s\n", cur, check, label)
	}

	if m.err != "" {
		fmt.Fprintf(&b, "  %s\n", errorStyle.Render("✗ "+m.err))
	}
	fmt.Fprintf(&b, "\n  %s\n", hintStyle.Render("↑/↓ navigate · space toggle · enter confirm · esc cancel"))
	return tea.NewView(b.String())
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
