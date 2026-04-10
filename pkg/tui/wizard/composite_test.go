package wizard

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	"github.com/verda-cloud/verdagostack/pkg/tui/bubbletea"
)

func TestComposite_CtrlC_ProducesExit(t *testing.T) {
	resultCh := make(chan promptResult, 1)
	cfg := tui.ResolveSelectConfig(nil)
	prompt := bubbletea.NewSelectPrompt("Pick", []string{"a", "b"}, cfg)

	m := newCompositeModel(DefaultKeyBindings(), nil, resultCh)
	m.setPrompt(prompt)

	// Send Ctrl+C
	_, _ = m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	select {
	case r := <-resultCh:
		if r.action != ActionExit {
			t.Errorf("expected ActionExit, got %v", r.action)
		}
	default:
		t.Fatal("expected result on channel")
	}
}

func TestComposite_Enter_ForwardedToPrompt(t *testing.T) {
	resultCh := make(chan promptResult, 1)
	cfg := tui.ResolveSelectConfig(nil)
	prompt := bubbletea.NewSelectPrompt("Pick", []string{"a"}, cfg)

	m := newCompositeModel(DefaultKeyBindings(), nil, resultCh)
	m.setPrompt(prompt)

	// Enter on first item
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	select {
	case r := <-resultCh:
		if r.action != ActionNone {
			t.Errorf("expected ActionNone (success), got %v", r.action)
		}
		if r.value != 0 { // index 0
			t.Errorf("expected value 0, got %v", r.value)
		}
	default:
		t.Fatal("expected result on channel")
	}
}

func TestComposite_GoBackMsg_ProducesBack(t *testing.T) {
	resultCh := make(chan promptResult, 1)
	cfg := tui.ResolveSelectConfig(nil)
	prompt := bubbletea.NewSelectPrompt("Pick", []string{"a"}, cfg)

	m := newCompositeModel(DefaultKeyBindings(), nil, resultCh)
	m.setPrompt(prompt)

	// Esc with no filter → GoBackMsg
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = updated.(compositeModel)

	// The cmd from the prompt returns GoBackMsg — deliver it back
	if cmd != nil {
		msg := cmd()
		_, _ = m.Update(msg)
	}

	select {
	case r := <-resultCh:
		if r.action != ActionBack {
			t.Errorf("expected ActionBack, got %v", r.action)
		}
	default:
		t.Fatal("expected result on channel")
	}
}

func TestComposite_View_ContainsPrompt(t *testing.T) {
	resultCh := make(chan promptResult, 1)
	cfg := tui.ResolveSelectConfig(nil)
	prompt := bubbletea.NewSelectPrompt("Pick color", []string{"red", "blue"}, cfg)

	m := newCompositeModel(DefaultKeyBindings(), nil, resultCh)
	m.setPrompt(prompt)

	view := m.View()
	if view.Content == "" {
		t.Fatal("expected non-empty view")
	}
}

func TestComposite_ShowPromptMsg_SwapsPrompt(t *testing.T) {
	resultCh := make(chan promptResult, 1)

	m := newCompositeModel(DefaultKeyBindings(), nil, resultCh)

	cfg := tui.ResolveSelectConfig(nil)
	prompt := bubbletea.NewSelectPrompt("New prompt", []string{"x"}, cfg)

	updated, _ := m.Update(showPromptMsg{model: prompt})
	m = updated.(compositeModel)

	if m.prompt == nil {
		t.Fatal("expected prompt to be set after showPromptMsg")
	}
}
