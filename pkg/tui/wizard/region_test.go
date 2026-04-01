package wizard

import (
	"reflect"
	"strings"
	"testing"
)

func TestStepChangedMsg_Fields(t *testing.T) {
	msg := StepChangedMsg{
		Current:   3,
		Total:     12,
		StepName:  "instance-type",
		Collected: map[string]any{"region": "FIN-01"},
	}
	if msg.Current != 3 || msg.Total != 12 {
		t.Error("StepChangedMsg fields not set correctly")
	}
	if msg.StepName != "instance-type" {
		t.Errorf("expected StepName 'instance-type', got %q", msg.StepName)
	}
}

func TestRegionDef_HasID(t *testing.T) {
	def := RegionDef{
		ID: "progress",
	}
	if def.ID != "progress" {
		t.Errorf("expected ID 'progress', got %q", def.ID)
	}
}

func TestSubscribeFilter(t *testing.T) {
	subs := []reflect.Type{reflect.TypeFor[StepChangedMsg]()}
	msgType := reflect.TypeFor[StepChangedMsg]()

	found := false
	for _, s := range subs {
		if s == msgType {
			found = true
		}
	}
	if !found {
		t.Error("StepChangedMsg should match subscription")
	}
}

func TestProgressRegion_Render(t *testing.T) {
	r := NewProgressRegion()

	out, pub := r.Update(StepChangedMsg{Current: 2, Total: 5, StepName: "gpu"})

	if pub != nil {
		t.Error("progress region should not publish messages")
	}
	if !strings.Contains(out, "40%") {
		t.Errorf("expected '40%%' in output, got: %s", out)
	}
}

func TestProgressRegion_SingleStepHidden(t *testing.T) {
	r := NewProgressRegion()

	out, _ := r.Update(StepChangedMsg{Current: 1, Total: 1, StepName: "only"})

	if out != "" {
		t.Errorf("single-step should produce empty output, got: %s", out)
	}
}

func TestProgressRegion_CustomGradient(t *testing.T) {
	r := NewProgressRegion(
		WithProgressGradient("#bd93f9", "#ff79c6"),
		WithProgressWidth(20),
	)

	out, _ := r.Update(StepChangedMsg{Current: 3, Total: 6, StepName: "gpu"})

	if !strings.Contains(out, "50%") {
		t.Errorf("expected '50%%' in output, got: %s", out)
	}
}

func TestProgressRegion_SolidFill(t *testing.T) {
	r := NewProgressRegion(WithProgressSolidFill("#50fa7b"))

	out, _ := r.Update(StepChangedMsg{Current: 1, Total: 3, StepName: "a"})

	if !strings.Contains(out, "33%") {
		t.Errorf("expected '33%%' in output, got: %s", out)
	}
}

func TestProgressRegion_StepLabel(t *testing.T) {
	r := NewProgressRegion(WithProgressStepLabel(), WithoutProgressPercent())

	out, _ := r.Update(StepChangedMsg{Current: 2, Total: 5, StepName: "gpu"})

	if !strings.Contains(out, "Step 2 of 5") {
		t.Errorf("expected 'Step 2 of 5', got: %s", out)
	}
}

func TestProgressRegion_IgnoresOtherMessages(t *testing.T) {
	r := NewProgressRegion()

	out, _ := r.Update(CollectedChangedMsg{Key: "x", Value: "y"})

	if out != "" {
		t.Errorf("should ignore non-StepChanged messages, got: %s", out)
	}
}

func TestCustomRegion_ReactsToCollectedChange(t *testing.T) {
	cost := &costRegion{}

	bus := NewMessageBus()
	bus.Register("cost", cost)

	bus.Broadcast(CollectedChangedMsg{
		Key:   "instance-type",
		Value: "1H100.80S",
		Collected: map[string]any{
			"instance-type": "1H100.80S",
		},
	})

	renders := bus.RenderAll()
	if !strings.Contains(renders[0], "$3.20/hr") {
		t.Errorf("expected cost display, got: %s", renders[0])
	}
}

type costRegion struct {
	last string
}

func (r *costRegion) Update(msg any) (string, []any) {
	if m, ok := msg.(CollectedChangedMsg); ok {
		if m.Collected["instance-type"] == "1H100.80S" {
			r.last = "  Estimated cost: $3.20/hr"
		}
	}
	return r.last, nil
}

func (r *costRegion) Subscribe() []reflect.Type {
	return nil
}
