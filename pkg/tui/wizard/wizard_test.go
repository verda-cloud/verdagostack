package wizard

import (
	"context"
	"testing"
)

func TestPromptType_Constants(t *testing.T) {
	types := []PromptType{
		SelectPrompt,
		MultiSelectPrompt,
		TextInputPrompt,
		ConfirmPrompt,
		PasswordPrompt,
	}
	seen := make(map[PromptType]bool)
	for _, pt := range types {
		if seen[pt] {
			t.Errorf("duplicate PromptType value: %d", pt)
		}
		seen[pt] = true
	}
}

func TestChoice_Fields(t *testing.T) {
	c := Choice{
		Label:       "H100 80GB - $3.20/hr",
		Value:       "1H100.80S.30V",
		Description: "8x NVIDIA H100",
	}
	if c.Label == "" || c.Value == "" {
		t.Error("Choice fields should be populated")
	}
}

func TestStep_Fields(t *testing.T) {
	s := Step{
		Name:        "gpu",
		Description: "Select GPU type",
		Prompt:      SelectPrompt,
		Required:    true,
		DependsOn:   []string{"location"},
	}
	if s.Name != "gpu" {
		t.Errorf("expected name 'gpu', got %q", s.Name)
	}
	if len(s.DependsOn) != 1 || s.DependsOn[0] != "location" {
		t.Error("DependsOn should contain 'location'")
	}
}

func TestFlow_Fields(t *testing.T) {
	f := Flow{
		Name:  "vm-create",
		Steps: []Step{{Name: "a"}, {Name: "b"}},
	}
	if f.Name != "vm-create" {
		t.Errorf("expected flow name 'vm-create', got %q", f.Name)
	}
	if len(f.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(f.Steps))
	}
}

func TestStaticChoices(t *testing.T) {
	loader := StaticChoices(
		Choice{Label: "A", Value: "a"},
		Choice{Label: "B", Value: "b"},
	)
	choices, err := loader(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(choices))
	}
	if choices[0].Value != "a" || choices[1].Value != "b" {
		t.Error("unexpected choice values")
	}
}
