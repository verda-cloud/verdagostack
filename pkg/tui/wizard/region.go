package wizard

import "reflect"

// Region is an actor that receives messages and renders output.
// Each region maintains its own state and renders independently.
type Region interface {
	// Update receives a message and returns:
	// - render: the new display string for this region
	// - publish: optional messages to broadcast to other regions
	Update(msg any) (render string, publish []any)

	// Subscribe returns the message types this region listens to.
	// nil means receive all engine broadcasts only.
	// Non-nil means receive only those types (plus engine broadcasts).
	Subscribe() []reflect.Type
}

// RegionDef defines a region in the layout.
type RegionDef struct {
	ID     string
	Region Region
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
