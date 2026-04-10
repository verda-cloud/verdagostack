package bubbletea

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPasswordPrompt_Hints(t *testing.T) {
	m := NewPasswordPrompt("Secret:")
	hints := m.Hints()
	if len(hints) == 0 {
		t.Fatal("expected hints")
	}
}

func TestPasswordPrompt_TypeAndEnter(t *testing.T) {
	m := NewPasswordPrompt("Secret:")

	// Type "pass"
	for _, ch := range "pass" {
		updated, _ := m.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
		m = updated.(PromptModel)
	}

	// Enter
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(PromptModel)

	val, done := m.Result()
	if !done {
		t.Fatal("expected done after Enter")
	}
	if val != "pass" {
		t.Errorf("expected 'pass', got %v", val)
	}
}

func TestPasswordPrompt_EscGoBack(t *testing.T) {
	m := NewPasswordPrompt("Secret:")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(GoBackMsg); !ok {
		t.Fatalf("expected GoBackMsg, got %T", msg)
	}
}

func TestPasswordPrompt_Result_NotDoneInitially(t *testing.T) {
	m := NewPasswordPrompt("Secret:")
	_, done := m.Result()
	if done {
		t.Fatal("should not be done initially")
	}
}
