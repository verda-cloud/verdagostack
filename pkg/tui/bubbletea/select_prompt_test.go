package bubbletea

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

func TestSelectPrompt_Hints(t *testing.T) {
	cfg := tui.ResolveSelectConfig(nil)
	m := NewSelectPrompt("Pick one", []string{"a", "b"}, cfg)
	hints := m.Hints()
	if len(hints) == 0 {
		t.Fatal("expected hints")
	}
}

func TestSelectPrompt_ArrowAndEnter(t *testing.T) {
	cfg := tui.ResolveSelectConfig(nil)
	m := NewSelectPrompt("Pick", []string{"alpha", "beta"}, cfg)

	// Move down
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(PromptModel)

	// Press Enter
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(PromptModel)

	val, done := m.Result()
	if !done {
		t.Fatal("expected done after Enter")
	}
	if val != 1 { // index 1 = "beta"
		t.Errorf("expected index 1, got %v", val)
	}
}

func TestSelectPrompt_EscWithFilter_ClearsFilter(t *testing.T) {
	cfg := tui.ResolveSelectConfig(nil)
	m := NewSelectPrompt("Pick", []string{"alpha", "beta"}, cfg)

	// Type to filter
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	m = updated.(PromptModel)

	// Esc should clear filter, not go back
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	_ = updated.(PromptModel)

	// Should NOT produce GoBackMsg
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(GoBackMsg); ok {
			t.Fatal("Esc with active filter should clear filter, not go back")
		}
	}
}

func TestSelectPrompt_EscWithoutFilter_GoBack(t *testing.T) {
	cfg := tui.ResolveSelectConfig(nil)
	m := NewSelectPrompt("Pick", []string{"alpha", "beta"}, cfg)

	// Esc with no filter should produce GoBackMsg
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(GoBackMsg); !ok {
		t.Fatalf("expected GoBackMsg, got %T", msg)
	}
}

func TestSelectPrompt_CtrlC_NotHandled(t *testing.T) {
	cfg := tui.ResolveSelectConfig(nil)
	m := NewSelectPrompt("Pick", []string{"a", "b"}, cfg)

	// Ctrl+C should NOT be handled — composite intercepts it
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	pm := updated.(PromptModel)
	_, done := pm.Result()
	if done {
		t.Fatal("Ctrl+C should not complete the prompt")
	}
	// In non-wizard path, Ctrl+C sets interrupted and returns tea.Quit.
	// The composite intercepts Ctrl+C before the model sees it, so this
	// test verifies the model doesn't mark as "done" via Result().
	_ = cmd
}

func TestSelectPrompt_Result_NotDoneInitially(t *testing.T) {
	cfg := tui.ResolveSelectConfig(nil)
	m := NewSelectPrompt("Pick", []string{"a"}, cfg)
	_, done := m.Result()
	if done {
		t.Fatal("should not be done initially")
	}
}
