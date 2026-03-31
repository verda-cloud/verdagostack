package wizard

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	tuitesting "github.com/verda-cloud/verdagostack/pkg/tui/testing"
)

func TestEngine_HappyPath_AllStepsPrompted(t *testing.T) {
	var region, gpu string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:        "region",
				Description: "Select region",
				Prompt:      SelectPrompt,
				Required:    true,
				Loader: StaticChoices(
					Choice{Label: "Finland", Value: "FIN-01"},
					Choice{Label: "Sweden", Value: "SWE-01"},
				),
				Setter: func(v any) { region = v.(string) },
			},
			{
				Name:        "gpu",
				Description: "Select GPU",
				Prompt:      SelectPrompt,
				Required:    true,
				DependsOn:   []string{"region"},
				Loader: func(_ context.Context, _ tui.Prompter, collected map[string]any) ([]Choice, error) {
					if collected["region"] == "FIN-01" {
						return []Choice{{Label: "H100", Value: "h100"}}, nil
					}
					return []Choice{{Label: "A100", Value: "a100"}}, nil
				},
				Setter: func(v any) { gpu = v.(string) },
			},
		},
	}

	p := tuitesting.New().AddSelect(0).AddSelect(0)
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if region != "FIN-01" {
		t.Errorf("expected region 'FIN-01', got %q", region)
	}
	if gpu != "h100" {
		t.Errorf("expected gpu 'h100', got %q", gpu)
	}
}

func TestEngine_SkipAlreadySet(t *testing.T) {
	var hostname string
	alreadySet := true

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "hostname",
				Prompt:   TextInputPrompt,
				Required: true,
				IsSet:    func() bool { return alreadySet },
				Setter:   func(v any) { hostname = v.(string) },
			},
		},
	}

	p := tuitesting.New()
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname != "" {
		t.Error("hostname should not be set — step was skipped")
	}
}

func TestEngine_ShouldSkip(t *testing.T) {
	var contract string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "category",
				Prompt:   SelectPrompt,
				Required: true,
				Loader:   StaticChoices(Choice{Label: "Spot", Value: "spot"}),
				Setter:   func(v any) {},
			},
			{
				Name:       "contract",
				Prompt:     SelectPrompt,
				Required:   false,
				ShouldSkip: func(c map[string]any) bool { return c["category"] == "spot" },
				Loader:     StaticChoices(Choice{Label: "Monthly", Value: "monthly"}),
				Setter:     func(v any) { contract = v.(string) },
			},
		},
	}

	p := tuitesting.New().AddSelect(0)
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contract != "" {
		t.Error("contract should be empty — step was skipped for spot")
	}
}

func TestEngine_TextInput(t *testing.T) {
	var hostname string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "hostname",
				Prompt:   TextInputPrompt,
				Required: true,
				Default:  func(_ map[string]any) any { return "my-vm-001" },
				Setter:   func(v any) { hostname = v.(string) },
			},
		},
	}

	p := tuitesting.New().AddTextInput("custom-host")
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname != "custom-host" {
		t.Errorf("expected 'custom-host', got %q", hostname)
	}
}

func TestEngine_MultiSelect(t *testing.T) {
	var keys []string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "ssh-keys",
				Prompt:   MultiSelectPrompt,
				Required: true,
				Loader: StaticChoices(
					Choice{Label: "key-a", Value: "id-a"},
					Choice{Label: "key-b", Value: "id-b"},
					Choice{Label: "key-c", Value: "id-c"},
				),
				Setter: func(v any) { keys = v.([]string) },
			},
		},
	}

	p := tuitesting.New().AddMultiSelect([]int{0, 2})
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 || keys[0] != "id-a" || keys[1] != "id-c" {
		t.Errorf("expected [id-a, id-c], got %v", keys)
	}
}

func TestEngine_Confirm(t *testing.T) {
	var isSpot bool

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "spot",
				Prompt:   ConfirmPrompt,
				Required: true,
				Setter:   func(v any) { isSpot = v.(bool) },
			},
		},
	}

	p := tuitesting.New().AddConfirm(true)
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isSpot {
		t.Error("expected isSpot to be true")
	}
}

func TestEngine_Password(t *testing.T) {
	var token string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "token",
				Prompt:   PasswordPrompt,
				Required: true,
				Setter:   func(v any) { token = v.(string) },
			},
		},
	}

	p := tuitesting.New().AddPassword("secret-123")
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "secret-123" {
		t.Errorf("expected 'secret-123', got %q", token)
	}
}

func TestEngine_DefaultUsedForOptional(t *testing.T) {
	var desc string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "description",
				Prompt:   TextInputPrompt,
				Required: false,
				Default:  func(_ map[string]any) any { return "auto-generated" },
				Setter:   func(v any) { desc = v.(string) },
			},
		},
	}

	p := tuitesting.New().AddTextInput("")
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "auto-generated" {
		t.Errorf("expected 'auto-generated', got %q", desc)
	}
}

func TestEngine_BackNavigation_EmptyRequired(t *testing.T) {
	var region, gpu string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "region",
				Prompt:   SelectPrompt,
				Required: true,
				Loader: StaticChoices(
					Choice{Label: "Finland", Value: "FIN-01"},
					Choice{Label: "Sweden", Value: "SWE-01"},
				),
				Setter: func(v any) { region = v.(string) },
			},
			{
				Name:      "gpu",
				Prompt:    SelectPrompt,
				Required:  true,
				DependsOn: []string{"region"},
				Loader: func(_ context.Context, _ tui.Prompter, collected map[string]any) ([]Choice, error) {
					if collected["region"] == "FIN-01" {
						return []Choice{}, nil // empty — triggers auto-back
					}
					return []Choice{{Label: "A100", Value: "a100"}}, nil
				},
				Setter: func(v any) { gpu = v.(string) },
			},
		},
	}

	// First: select FIN-01 (idx 0) → gpu empty → auto-back
	// Second: select SWE-01 (idx 1) → gpu has A100 → select idx 0
	p := tuitesting.New().AddSelect(0).AddSelect(1).AddSelect(0)
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if region != "SWE-01" {
		t.Errorf("expected region 'SWE-01', got %q", region)
	}
	if gpu != "a100" {
		t.Errorf("expected gpu 'a100', got %q", gpu)
	}
}

func TestEngine_EmptyRequired_AtFirstStep_ReturnsError(t *testing.T) {
	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "gpu",
				Prompt:   SelectPrompt,
				Required: true,
				Loader: func(_ context.Context, _ tui.Prompter, _ map[string]any) ([]Choice, error) {
					return []Choice{}, nil
				},
				Setter: func(v any) {},
			},
		},
	}

	p := tuitesting.New()
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err == nil {
		t.Fatal("expected error when first step has no options")
	}
}

func TestEngine_ValidationError(t *testing.T) {
	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "size",
				Prompt:   TextInputPrompt,
				Required: true,
				Validate: func(v any) error {
					if v.(string) == "" {
						return fmt.Errorf("size cannot be empty")
					}
					return nil
				},
				Setter: func(v any) {},
			},
		},
	}

	p := tuitesting.New().AddTextInput("")
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "size cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEngine_CachesLoaderResults(t *testing.T) {
	loadCount := 0

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "static",
				Prompt:   SelectPrompt,
				Required: true,
				Loader: func(_ context.Context, _ tui.Prompter, _ map[string]any) ([]Choice, error) {
					loadCount++
					return []Choice{{Label: "A", Value: "a"}}, nil
				},
				Setter: func(v any) {},
			},
		},
	}

	p := tuitesting.New().AddSelect(0)
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loadCount != 1 {
		t.Errorf("expected loader called once, got %d", loadCount)
	}
}
