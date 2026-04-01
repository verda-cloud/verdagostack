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

func TestSelectModel_Refilter(t *testing.T) {
	choices := []string{"Apple", "Apricot", "Banana", "Blueberry", "Cherry"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	// Filter by "ap" — should match Apple (0), Apricot (1) (case-insensitive)
	mp := &m
	mp.filter = "ap"
	mp.refilter()

	if len(mp.matched) != 2 {
		t.Fatalf("expected 2 matches, got %d: %v", len(mp.matched), mp.matched)
	}
	if mp.matched[0] != 0 || mp.matched[1] != 1 {
		t.Errorf("matched = %v, want [0 1]", mp.matched)
	}
	if mp.cursor != 0 {
		t.Errorf("cursor should reset to 0, got %d", mp.cursor)
	}
}

func TestSelectModel_RefilterNoMatch(t *testing.T) {
	choices := []string{"Apple", "Banana", "Cherry"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	mp := &m
	mp.filter = "zzz"
	mp.refilter()

	if len(mp.matched) != 0 {
		t.Errorf("expected 0 matches, got %d", len(mp.matched))
	}
	if mp.cursor != 0 {
		t.Errorf("cursor should be 0, got %d", mp.cursor)
	}
}

func TestSelectModel_RefilterEmpty(t *testing.T) {
	choices := []string{"Apple", "Banana", "Cherry"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	mp := &m
	mp.filter = "ap"
	mp.refilter()
	// Clear filter
	mp.filter = ""
	mp.refilter()

	if len(mp.matched) != 3 {
		t.Errorf("expected all 3 choices back, got %d", len(mp.matched))
	}
}
