package bubbletea

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

func TestTextInputPrompt_Hints(t *testing.T) {
	cfg := tui.ResolveTextInputConfig(nil)
	m := NewTextInputPrompt("Name:", cfg)
	hints := m.Hints()
	if len(hints) == 0 {
		t.Fatal("expected hints")
	}
}

func TestTextInputPrompt_TypeAndEnter(t *testing.T) {
	cfg := tui.ResolveTextInputConfig(nil)
	m := NewTextInputPrompt("Name:", cfg)

	// Type "hello"
	for _, ch := range "hello" {
		updated, _ := m.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
		m = updated.(PromptModel)
	}

	// Press Enter
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(PromptModel)

	val, done := m.Result()
	if !done {
		t.Fatal("expected done after Enter")
	}
	if val != "hello" {
		t.Errorf("expected 'hello', got %v", val)
	}
}

func TestTextInputPrompt_EscGoBack(t *testing.T) {
	cfg := tui.ResolveTextInputConfig(nil)
	m := NewTextInputPrompt("Name:", cfg)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(GoBackMsg); !ok {
		t.Fatalf("expected GoBackMsg, got %T", msg)
	}
}

func TestTextInputPrompt_Result_NotDoneInitially(t *testing.T) {
	cfg := tui.ResolveTextInputConfig(nil)
	m := NewTextInputPrompt("Name:", cfg)
	_, done := m.Result()
	if done {
		t.Fatal("should not be done initially")
	}
}
