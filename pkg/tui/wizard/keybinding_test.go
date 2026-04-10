package wizard

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestKeyBinding_MatchCtrlC(t *testing.T) {
	bindings := DefaultKeyBindings()
	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}

	action, ok := MatchBinding(bindings, msg)
	if !ok {
		t.Fatal("expected Ctrl+C to match a binding")
	}
	if action != ActionExit {
		t.Errorf("expected ActionExit, got %v", action)
	}
}

func TestKeyBinding_NoMatchForEnter(t *testing.T) {
	bindings := DefaultKeyBindings()
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}

	_, ok := MatchBinding(bindings, msg)
	if ok {
		t.Fatal("Enter should not match any wizard binding")
	}
}

func TestKeyBinding_NoMatchForEsc(t *testing.T) {
	bindings := DefaultKeyBindings()
	msg := tea.KeyPressMsg{Code: tea.KeyEscape}

	_, ok := MatchBinding(bindings, msg)
	if ok {
		t.Fatal("Esc should not match any wizard binding — it goes to the prompt")
	}
}

func TestKeyBinding_CustomBinding(t *testing.T) {
	bindings := []KeyBinding{
		{Key: KeyPattern{Code: 'q', Mod: tea.ModCtrl}, Action: ActionExit, Label: "ctrl+q exit"},
	}
	msg := tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}

	action, ok := MatchBinding(bindings, msg)
	if !ok {
		t.Fatal("expected Ctrl+Q to match custom binding")
	}
	if action != ActionExit {
		t.Errorf("expected ActionExit, got %v", action)
	}
}
