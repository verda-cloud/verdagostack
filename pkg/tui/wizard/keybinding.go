package wizard

import tea "charm.land/bubbletea/v2"

// Action represents a wizard-level command triggered by a key binding.
type Action int

const (
	ActionExit Action = iota // exit the wizard
	ActionBack               // go to previous step (reserved for future use)
)

// KeyPattern matches a tea.KeyPressMsg.
type KeyPattern struct {
	Code rune
	Mod  tea.KeyMod
}

// KeyBinding maps a key pattern to a wizard-level action.
type KeyBinding struct {
	Key    KeyPattern
	Action Action
	Label  string // displayed in hint bar
}

// DefaultKeyBindings returns the default wizard key bindings.
func DefaultKeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: KeyPattern{Code: 'c', Mod: tea.ModCtrl}, Action: ActionExit, Label: "ctrl+c exit"},
	}
}

// MatchBinding checks if a key message matches any binding.
// Returns the action and true if matched, or zero and false if not.
func MatchBinding(bindings []KeyBinding, msg tea.KeyPressMsg) (Action, bool) {
	for _, b := range bindings {
		if msg.Code == b.Key.Code && msg.Mod == b.Key.Mod {
			return b.Action, true
		}
	}
	return 0, false
}
