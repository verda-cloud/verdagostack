package bubbletea

import (
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

func TestNewSelectModel_InitialFilterState(t *testing.T) {
	choices := []string{"Apple", "Banana", "Cherry"}
	m := newSelectModel("Pick fruit", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	if m.filter != "" {
		t.Errorf("expected empty filter, got %q", m.filter)
	}
	if len(m.matched) != len(choices) {
		t.Errorf("expected matched length %d, got %d", len(choices), len(m.matched))
	}
	for i, idx := range m.matched {
		if idx != i {
			t.Errorf("matched[%d] = %d, want %d", i, idx, i)
		}
	}
}
