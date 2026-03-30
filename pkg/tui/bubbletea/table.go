package bubbletea

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

// Table implements tui.Status.
func (p *Prompter) Table(_ context.Context, columns []string, rows [][]string, opts ...tui.TableOption) error {
	cfg := tui.ResolveTableConfig(opts)

	cols := make([]table.Column, len(columns))
	for i, c := range columns {
		width := len(c)
		for _, row := range rows {
			if i < len(row) && len(row[i]) > width {
				width = len(row[i])
			}
		}
		if cfg.MaxWidth > 0 && width > cfg.MaxWidth/len(columns) {
			width = cfg.MaxWidth / len(columns)
		}
		cols[i] = table.Column{Title: c, Width: width}
	}

	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = row
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(tableRows),
		table.WithHeight(len(rows)+1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	_, err := fmt.Fprintln(p.out, t.View())
	return err
}
