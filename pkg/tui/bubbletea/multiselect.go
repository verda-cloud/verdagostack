package bubbletea

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

type multiSelectModel struct {
	prompt      string
	choices     []string
	filter      string
	matched     []int // indices into choices that match filter
	cursor      int   // position within matched (not choices)
	selected    map[int]bool
	pageSize    int
	loop        bool
	min         int
	max         int
	done        bool
	aborted     bool
	interrupted bool // true for Ctrl+C (hard cancel), false for Esc (soft cancel)
	err         string
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
	matched := make([]int, len(choices))
	for i := range choices {
		matched[i] = i
	}
	return multiSelectModel{
		prompt:   prompt,
		choices:  choices,
		matched:  matched,
		selected: selected,
		pageSize: ps,
		loop:     cfg.Loop,
		min:      cfg.Min,
		max:      cfg.Max,
	}
}

func (m *multiSelectModel) refilter() {
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

func (m *multiSelectModel) moveUp() {
	if len(m.matched) == 0 {
		return
	}
	if m.cursor > 0 {
		m.cursor--
	} else if m.loop {
		m.cursor = len(m.matched) - 1
	}
}

func (m *multiSelectModel) moveDown() {
	if len(m.matched) == 0 {
		return
	}
	if m.cursor < len(m.matched)-1 {
		m.cursor++
	} else if m.loop {
		m.cursor = 0
	}
}

// toggleAll selects or deselects all matched (visible) items.
func (m *multiSelectModel) toggleAll() {
	allSelected := true
	for _, idx := range m.matched {
		if !m.selected[idx] {
			allSelected = false
			break
		}
	}
	if allSelected {
		for _, idx := range m.matched {
			delete(m.selected, idx)
		}
	} else {
		for _, idx := range m.matched {
			if m.max > 0 && len(m.selected) >= m.max {
				m.err = fmt.Sprintf("maximum %d selections allowed", m.max)
				return
			}
			m.selected[idx] = true
		}
	}
	m.err = ""
}

func (m multiSelectModel) Init() tea.Cmd { return nil }

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp:
			m.moveUp()
		case tea.KeyDown:
			m.moveDown()
		case tea.KeySpace:
			if len(m.matched) == 0 {
				return m, nil
			}
			idx := m.matched[m.cursor]
			if m.selected[idx] {
				delete(m.selected, idx)
			} else {
				if m.max > 0 && len(m.selected) >= m.max {
					m.err = fmt.Sprintf("maximum %d selections allowed", m.max)
				} else {
					m.selected[idx] = true
					m.err = ""
				}
			}
		case tea.KeyEnter:
			if m.min > 0 && len(m.selected) < m.min {
				m.err = fmt.Sprintf("at least %d selections required", m.min)
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case tea.KeyEscape:
			if m.filter != "" {
				m.filter = ""
				m.refilter()
			} else {
				m.aborted = true
				return m, func() tea.Msg { return GoBackMsg{} }
			}
		case tea.KeyBackspace:
			if len(m.filter) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.filter)
				m.filter = m.filter[:len(m.filter)-size]
				m.refilter()
			}
		default:
			if msg.Mod&tea.ModCtrl != 0 {
				switch msg.Code {
				case 'a':
					m.toggleAll()
					return m, nil
				case 'c':
					m.interrupted = true
					return m, tea.Quit
				}
			}
			if msg.Text != "" {
				s := msg.Text
				// j/k navigation only when not filtering.
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
	}
	return m, nil
}

// Hints returns multiselect-specific key hints for the hint bar.
func (m multiSelectModel) Hints() []string {
	return []string{"↑/↓ navigate", "space toggle", "ctrl+a select all", "type to filter", "enter confirm", "esc back"}
}

// Result returns the selected indices after the user confirms.
func (m multiSelectModel) Result() (any, bool) {
	if !m.done {
		return nil, false
	}
	var indices []int
	for i := range m.choices {
		if m.selected[i] {
			indices = append(indices, i)
		}
	}
	return indices, true
}

// NewMultiSelectPrompt creates a multiselect prompt model for use in the wizard composite.
func NewMultiSelectPrompt(prompt string, choices []string, cfg tui.MultiSelectConfig) PromptModel {
	return newMultiSelectModel(prompt, choices, cfg)
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
	fmt.Fprintf(&b, "%s %s", promptStyle.Render("?"), titleStyle.Render(m.prompt))
	if m.filter != "" {
		fmt.Fprintf(&b, " %s", answerStyle.Render(m.filter))
	}
	b.WriteString("\n")

	if len(m.matched) == 0 {
		fmt.Fprintf(&b, "    %s\n", dimStyle.Render("no matches"))
	} else {
		start, end := m.visibleRange()
		for i := start; i < end; i++ {
			idx := m.matched[i]
			cur := "  "
			if i == m.cursor {
				cur = cursorStyle.Render("> ")
			}
			check := uncheckStyle.Render("[ ]")
			label := dimStyle.Render(m.choices[idx])
			if m.selected[idx] {
				check = checkStyle.Render("[x]")
				label = selectedStyle.Render(m.choices[idx])
			}
			if i == m.cursor && !m.selected[idx] {
				label = selectedStyle.Render(m.choices[idx])
			}
			fmt.Fprintf(&b, "  %s%s %s\n", cur, check, label)
		}
	}

	if m.err != "" {
		fmt.Fprintf(&b, "  %s\n", errorStyle.Render("✗ "+m.err))
	}
	return tea.NewView(b.String())
}

func (m multiSelectModel) visibleRange() (int, int) {
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

// MultiSelect implements tui.Prompter.
func (p *Prompter) MultiSelect(ctx context.Context, prompt string, choices []string, opts ...tui.MultiSelectOption) ([]int, error) {
	if len(choices) == 0 {
		return nil, fmt.Errorf("multi-select prompt: no choices provided")
	}

	cfg := tui.ResolveMultiSelectConfig(opts)
	model := newMultiSelectModel(prompt, choices, cfg)

	r := p.runProgram(ctx, model)
	if r.interrupted {
		return nil, tui.ErrInterrupted
	}
	if r.err != nil {
		return nil, fmt.Errorf("multi-select prompt: %w", r.err)
	}

	m := r.model.(multiSelectModel)
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
