package bubbletea

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

func TestConfirmPrompt_Hints(t *testing.T) {
	cfg := tui.ResolveConfirmConfig(nil)
	m := NewConfirmPrompt("Continue?", cfg)
	hints := m.Hints()
	if len(hints) == 0 {
		t.Fatal("expected hints")
	}
}

func TestConfirmPrompt_YesKey(t *testing.T) {
	cfg := tui.ResolveConfirmConfig(nil)
	m := NewConfirmPrompt("Continue?", cfg)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	m = updated.(PromptModel)

	val, done := m.Result()
	if !done {
		t.Fatal("expected done after y")
	}
	if val != true {
		t.Errorf("expected true, got %v", val)
	}
}

func TestConfirmPrompt_NoKey(t *testing.T) {
	cfg := tui.ResolveConfirmConfig(nil)
	m := NewConfirmPrompt("Continue?", cfg)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	m = updated.(PromptModel)

	val, done := m.Result()
	if !done {
		t.Fatal("expected done after n")
	}
	if val != false {
		t.Errorf("expected false, got %v", val)
	}
}

func TestConfirmPrompt_EnterUsesDefault(t *testing.T) {
	cfg := tui.ResolveConfirmConfig([]tui.ConfirmOption{tui.WithConfirmDefault(true)})
	m := NewConfirmPrompt("Continue?", cfg)

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(PromptModel)

	val, done := m.Result()
	if !done {
		t.Fatal("expected done after Enter")
	}
	if val != true {
		t.Errorf("expected true (default), got %v", val)
	}
}

func TestConfirmPrompt_EscGoBack(t *testing.T) {
	cfg := tui.ResolveConfirmConfig(nil)
	m := NewConfirmPrompt("Continue?", cfg)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(GoBackMsg); !ok {
		t.Fatalf("expected GoBackMsg, got %T", msg)
	}
}

func TestConfirmPrompt_Result_NotDoneInitially(t *testing.T) {
	cfg := tui.ResolveConfirmConfig(nil)
	m := NewConfirmPrompt("Continue?", cfg)
	_, done := m.Result()
	if done {
		t.Fatal("should not be done initially")
	}
}
