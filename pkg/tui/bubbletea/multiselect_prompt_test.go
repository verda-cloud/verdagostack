package bubbletea

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

func TestMultiSelectPrompt_Hints(t *testing.T) {
	cfg := tui.ResolveMultiSelectConfig(nil)
	m := NewMultiSelectPrompt("Pick", []string{"a", "b"}, cfg)
	hints := m.Hints()
	if len(hints) == 0 {
		t.Fatal("expected hints")
	}
}

func TestMultiSelectPrompt_SpaceAndEnter(t *testing.T) {
	cfg := tui.ResolveMultiSelectConfig(nil)
	m := NewMultiSelectPrompt("Pick", []string{"alpha", "beta", "gamma"}, cfg)

	// Space to toggle first item
	updated, _ := m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = updated.(PromptModel)

	// Down + Space to toggle second
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(PromptModel)
	updated, _ = m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = updated.(PromptModel)

	// Enter to confirm
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(PromptModel)

	val, done := m.Result()
	if !done {
		t.Fatal("expected done after Enter")
	}
	indices, ok := val.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", val)
	}
	if len(indices) != 2 || indices[0] != 0 || indices[1] != 1 {
		t.Errorf("expected [0 1], got %v", indices)
	}
}

func TestMultiSelectPrompt_EscGoBack(t *testing.T) {
	cfg := tui.ResolveMultiSelectConfig(nil)
	m := NewMultiSelectPrompt("Pick", []string{"a"}, cfg)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(GoBackMsg); !ok {
		t.Fatalf("expected GoBackMsg, got %T", msg)
	}
}

func TestMultiSelectPrompt_Result_NotDoneInitially(t *testing.T) {
	cfg := tui.ResolveMultiSelectConfig(nil)
	m := NewMultiSelectPrompt("Pick", []string{"a"}, cfg)
	_, done := m.Result()
	if done {
		t.Fatal("should not be done initially")
	}
}
