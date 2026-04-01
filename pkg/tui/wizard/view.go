package wizard

import "reflect"

// View is an actor that receives messages and renders output.
// Each view maintains its own state and renders independently.
type View interface {
	// Update receives a message and returns:
	// - render: the new display string for this view
	// - publish: optional messages to broadcast to other views
	Update(msg any) (render string, publish []any)

	// Subscribe returns the message types this view listens to.
	// nil means receive all engine broadcasts only.
	// Non-nil means receive only those types (plus engine broadcasts).
	Subscribe() []reflect.Type
}

// ViewDef defines a view in the layout.
type ViewDef struct {
	ID   string
	View View
}

// --- Engine broadcast messages ---

// StepChangedMsg is broadcast when the engine moves to a new step.
type StepChangedMsg struct {
	Current   int
	Total     int
	StepName  string
	Collected map[string]any
}

// CollectedChangedMsg is broadcast when a step completes and collected values change.
type CollectedChangedMsg struct {
	Key       string
	Value     any
	Collected map[string]any
}

// StoreChangedMsg is broadcast when a value in the store is set.
type StoreChangedMsg struct {
	Key   string
	Value any
}
