package bubbletea

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

// --- Task 3: Typing/Backspace/Update tests ---

func sendKey(m selectModel, msg tea.KeyMsg) selectModel {
	result, _ := m.Update(msg)
	return result.(selectModel)
}

func sendRune(m selectModel, r rune) selectModel {
	return sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
}

func TestSelectModel_TypeToFilter(t *testing.T) {
	choices := []string{"Apple", "Apricot", "Banana", "Blueberry", "Cherry"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	m = sendRune(m, 'b')

	if m.filter != "b" {
		t.Errorf("filter = %q, want %q", m.filter, "b")
	}
	if len(m.matched) != 2 {
		t.Errorf("expected 2 matches, got %d", len(m.matched))
	}
}

func TestSelectModel_Backspace(t *testing.T) {
	choices := []string{"Apple", "Apricot", "Banana"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	m = sendRune(m, 'a')
	m = sendRune(m, 'p')

	if m.filter != "ap" {
		t.Fatalf("filter = %q, want %q", m.filter, "ap")
	}

	m = sendKey(m, tea.KeyMsg{Type: tea.KeyBackspace})

	if m.filter != "a" {
		t.Errorf("filter = %q after backspace, want %q", m.filter, "a")
	}
	if len(m.matched) != 3 {
		t.Errorf("expected 3 matches, got %d", len(m.matched))
	}
}

func TestSelectModel_VimKeysNavigateWithoutFilter(t *testing.T) {
	choices := []string{"Apple", "Banana", "Cherry"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	m = sendRune(m, 'j')
	if m.cursor != 1 {
		t.Errorf("j should move down: cursor = %d, want 1", m.cursor)
	}
	if m.filter != "" {
		t.Errorf("j without filter should not set filter, got %q", m.filter)
	}

	m = sendRune(m, 'k')
	if m.cursor != 0 {
		t.Errorf("k should move up: cursor = %d, want 0", m.cursor)
	}
}

func TestSelectModel_VimKeysTypeWhenFiltering(t *testing.T) {
	choices := []string{"Ajax", "Jolt", "Koji"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	// Start typing — first char is 'a', then 'j' should go into filter
	m = sendRune(m, 'a')
	m = sendRune(m, 'j')

	if m.filter != "aj" {
		t.Errorf("filter = %q, want %q", m.filter, "aj")
	}
}

// --- Task 4: View rendering tests ---

func TestSelectModel_ViewShowsFilter(t *testing.T) {
	choices := []string{"Apple", "Apricot", "Banana"}
	m := newSelectModel("Pick fruit", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	m = sendRune(m, 'a')
	m = sendRune(m, 'p')

	view := m.View()

	if !strings.Contains(view, "ap") {
		t.Errorf("view should show filter text 'ap', got:\n%s", view)
	}
	if !strings.Contains(view, "Apple") {
		t.Error("view should show Apple")
	}
	if !strings.Contains(view, "Apricot") {
		t.Error("view should show Apricot")
	}
	if strings.Contains(view, "Banana") {
		t.Error("view should NOT show Banana")
	}
}

func TestSelectModel_ViewNoMatchMessage(t *testing.T) {
	choices := []string{"Apple", "Banana"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	m = sendRune(m, 'z')

	view := m.View()
	if !strings.Contains(view, "no matches") {
		t.Errorf("view should show 'no matches' message, got:\n%s", view)
	}
}

// --- Task 5: Return value mapping test ---

func TestSelectModel_ChosenReturnsOriginalIndex(t *testing.T) {
	choices := []string{"Apple", "Apricot", "Banana", "Blueberry", "Cherry"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	m = sendRune(m, 'b')
	m = sendRune(m, 'l')

	if len(m.matched) != 1 {
		t.Fatalf("expected 1 match, got %d: %v", len(m.matched), m.matched)
	}

	m = sendKey(m, tea.KeyMsg{Type: tea.KeyEnter})

	if !m.chosen {
		t.Fatal("expected chosen=true")
	}
	originalIdx := m.matched[m.cursor]
	if originalIdx != 3 {
		t.Errorf("original index = %d, want 3 (Blueberry)", originalIdx)
	}
}

// --- Task 6: Edge case tests ---

func TestNewSelectModel_DefaultIndex(t *testing.T) {
	choices := []string{"Apple", "Banana", "Cherry"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{
		Default:  2,
		PageSize: 10,
		Loop:     true,
	})

	if m.cursor != 2 {
		t.Errorf("cursor = %d, want 2", m.cursor)
	}
}

func TestSelectModel_EscClearsFilterThenAborts(t *testing.T) {
	choices := []string{"Apple", "Banana"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	m = sendRune(m, 'a')

	// First Esc — clears filter
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyEscape})

	if m.filter != "" {
		t.Errorf("filter should be cleared, got %q", m.filter)
	}
	if m.aborted {
		t.Error("should not be aborted after first Esc")
	}
	if len(m.matched) != 2 {
		t.Errorf("all choices should be restored, got %d", len(m.matched))
	}

	// Second Esc — aborts
	m = sendKey(m, tea.KeyMsg{Type: tea.KeyEscape})

	if !m.aborted {
		t.Error("should be aborted after second Esc")
	}
}

func TestSelectModel_EnterWithNoMatchesDoesNothing(t *testing.T) {
	choices := []string{"Apple", "Banana"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	for _, r := range "zzz" {
		m = sendRune(m, r)
	}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(selectModel)

	if m.chosen {
		t.Error("should not be chosen when no matches")
	}
	if cmd != nil {
		t.Error("should not quit when no matches")
	}
}

func TestSelectModel_NavigationWrapsInFilteredList(t *testing.T) {
	choices := []string{"Apple", "Apricot", "Banana", "Avocado"}
	m := newSelectModel("Pick", choices, tui.SelectConfig{PageSize: 10, Loop: true})

	// "ap" matches Apple(0), Apricot(1) only
	m = sendRune(m, 'a')
	m = sendRune(m, 'p')

	if len(m.matched) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(m.matched))
	}

	down := tea.KeyMsg{Type: tea.KeyDown}
	m = sendKey(m, down)

	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}

	// Down again should wrap to 0
	m = sendKey(m, down)

	if m.cursor != 0 {
		t.Errorf("cursor = %d after wrap, want 0", m.cursor)
	}
}
