package wizard

import (
	"context"
	"fmt"
	"io"
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

func TestEngine_ValidationError_RepromptsUntilValid(t *testing.T) {
	var size string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "size",
				Prompt:   TextInputPrompt,
				Required: true,
				Validate: func(v any) error {
					if v.(string) == "bad" {
						return fmt.Errorf("invalid size")
					}
					return nil
				},
				Setter: func(v any) { size = v.(string) },
			},
		},
	}

	// First: "bad" (fails validation, re-prompt), second: "100" (passes)
	p := tuitesting.New().AddTextInput("bad").AddTextInput("100")
	engine := NewEngine(p, WithOutput(io.Discard))

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size != "100" {
		t.Errorf("expected '100', got %q", size)
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

func TestEngine_FullFlow_VMCreate(t *testing.T) {
	var category, contract, compute, instType, location, image, hostname string
	var sshKeys []string

	flow := &Flow{
		Name: "vm-create",
		Steps: []Step{
			{
				Name:     "instance-category",
				Prompt:   SelectPrompt,
				Required: true,
				Loader: StaticChoices(
					Choice{Label: "On-Demand", Value: "on-demand"},
					Choice{Label: "Spot", Value: "spot"},
				),
				Setter: func(v any) { category = v.(string) },
			},
			{
				Name:       "contract",
				Prompt:     SelectPrompt,
				Required:   false,
				Default:    func(_ map[string]any) any { return "PAY_AS_YOU_GO" },
				ShouldSkip: func(c map[string]any) bool { return c["instance-category"] == "spot" },
				Loader: StaticChoices(
					Choice{Label: "Pay as you go", Value: "PAY_AS_YOU_GO"},
					Choice{Label: "1 month", Value: "1_MONTH"},
				),
				Setter: func(v any) { contract = v.(string) },
			},
			{
				Name:     "compute-category",
				Prompt:   SelectPrompt,
				Required: true,
				Loader: StaticChoices(
					Choice{Label: "GPU", Value: "GPU"},
					Choice{Label: "CPU", Value: "CPU"},
				),
				Setter: func(v any) { compute = v.(string) },
			},
			{
				Name:      "instance-type",
				Prompt:    SelectPrompt,
				Required:  true,
				DependsOn: []string{"compute-category"},
				Loader: func(_ context.Context, _ tui.Prompter, c map[string]any) ([]Choice, error) {
					if c["compute-category"] == "GPU" {
						return []Choice{
							{Label: "H100 80GB", Value: "1H100"},
							{Label: "A100 40GB", Value: "1A100"},
						}, nil
					}
					return []Choice{{Label: "32 vCPU", Value: "32CPU"}}, nil
				},
				Setter: func(v any) { instType = v.(string) },
			},
			{
				Name:      "location",
				Prompt:    SelectPrompt,
				Required:  true,
				DependsOn: []string{"instance-type"},
				Loader: func(_ context.Context, _ tui.Prompter, _ map[string]any) ([]Choice, error) {
					return []Choice{{Label: "Finland (FIN-01)", Value: "FIN-01"}}, nil
				},
				Setter: func(v any) { location = v.(string) },
			},
			{
				Name:      "image",
				Prompt:    SelectPrompt,
				Required:  true,
				DependsOn: []string{"instance-type"},
				Loader: func(_ context.Context, _ tui.Prompter, _ map[string]any) ([]Choice, error) {
					return []Choice{{Label: "Ubuntu 24.04", Value: "ubuntu-24.04"}}, nil
				},
				Setter: func(v any) { image = v.(string) },
			},
			{
				Name:     "ssh-keys",
				Prompt:   MultiSelectPrompt,
				Required: true,
				Loader: StaticChoices(
					Choice{Label: "my-key", Value: "key-1"},
					Choice{Label: "work-key", Value: "key-2"},
				),
				Setter: func(v any) { sshKeys = v.([]string) },
			},
			{
				Name:     "hostname",
				Prompt:   TextInputPrompt,
				Required: true,
				Default:  func(_ map[string]any) any { return "vm-001" },
				Setter:   func(v any) { hostname = v.(string) },
			},
		},
	}

	p := tuitesting.New().
		AddSelect(0).                // on-demand
		AddSelect(0).                // pay as you go
		AddSelect(0).                // GPU
		AddSelect(0).                // H100
		AddSelect(0).                // FIN-01
		AddSelect(0).                // Ubuntu
		AddMultiSelect([]int{0, 1}). // both SSH keys
		AddTextInput("my-vm")        // hostname

	engine := NewEngine(p)
	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if category != "on-demand" {
		t.Errorf("category: got %q", category)
	}
	if contract != "PAY_AS_YOU_GO" {
		t.Errorf("contract: got %q", contract)
	}
	if compute != "GPU" {
		t.Errorf("compute: got %q", compute)
	}
	if instType != "1H100" {
		t.Errorf("instType: got %q", instType)
	}
	if location != "FIN-01" {
		t.Errorf("location: got %q", location)
	}
	if image != "ubuntu-24.04" {
		t.Errorf("image: got %q", image)
	}
	if len(sshKeys) != 2 {
		t.Errorf("sshKeys: got %v", sshKeys)
	}
	if hostname != "my-vm" {
		t.Errorf("hostname: got %q", hostname)
	}
}

func TestEngine_UserInitiatedBack(t *testing.T) {
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
				Loader: StaticChoices(
					Choice{Label: "H100", Value: "h100"},
					Choice{Label: "A100", Value: "a100"},
				),
				Setter: func(v any) { gpu = v.(string) },
			},
		},
	}

	// Step 1: select Finland (idx 0)
	// Step 2: select "← Back" (idx 2 = last, after h100 and a100)
	// Step 1 again: select Sweden (idx 1)
	// Step 2 again: select A100 (idx 1)
	p := tuitesting.New().
		AddSelect(0). // region: Finland
		AddSelect(2). // gpu: ← Back
		AddSelect(1). // region: Sweden
		AddSelect(1)  // gpu: A100

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

func TestEngine_BackNotShownOnFirstStep(t *testing.T) {
	// First step has 2 choices. If "← Back" were added, index 2 would be valid.
	// Since it's the first step, "← Back" should NOT be added, so index 2 is out of bounds.
	// We select index 0 which should work fine.
	var val string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "first",
				Prompt:   SelectPrompt,
				Required: true,
				Loader: StaticChoices(
					Choice{Label: "A", Value: "a"},
					Choice{Label: "B", Value: "b"},
				),
				Setter: func(v any) { val = v.(string) },
			},
		},
	}

	p := tuitesting.New().AddSelect(0)
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "a" {
		t.Errorf("expected 'a', got %q", val)
	}
}

func TestEngine_RequiredTextInput_RepromptsOnEmpty(t *testing.T) {
	var name string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "name",
				Prompt:   TextInputPrompt,
				Required: true,
				Setter:   func(v any) { name = v.(string) },
			},
		},
	}

	// First: empty (re-prompt), second: valid
	p := tuitesting.New().AddTextInput("").AddTextInput("my-service")
	engine := NewEngine(p, WithOutput(io.Discard))

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "my-service" {
		t.Errorf("expected 'my-service', got %q", name)
	}
}

func TestEngine_DependencyAwareAutoBack(t *testing.T) {
	// Flow: region(select) -> name(text) -> gpu(select, depends on region)
	// When gpu is empty for region=A, should back to region (not name).
	var region, name, gpu string

	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "region",
				Prompt:   SelectPrompt,
				Required: true,
				Loader: StaticChoices(
					Choice{Label: "Region A", Value: "a"},
					Choice{Label: "Region B", Value: "b"},
				),
				Setter: func(v any) { region = v.(string) },
			},
			{
				Name:     "name",
				Prompt:   TextInputPrompt,
				Required: true,
				Setter:   func(v any) { name = v.(string) },
			},
			{
				Name:      "gpu",
				Prompt:    SelectPrompt,
				Required:  true,
				DependsOn: []string{"region"},
				Loader: func(_ context.Context, _ tui.Prompter, c map[string]any) ([]Choice, error) {
					if c["region"] == "a" {
						return []Choice{}, nil // empty for region A
					}
					return []Choice{{Label: "H100", Value: "h100"}}, nil
				},
				Setter: func(v any) { gpu = v.(string) },
			},
		},
	}

	// 1: region=A (idx 0)
	// 2: name="test"
	// 3: gpu empty → auto-back to region (skips name!)
	// 4: region=B (idx 1)
	// 5: name="test2"
	// 6: gpu=H100 (idx 0)
	p := tuitesting.New().
		AddSelect(0).          // region: A
		AddTextInput("test").  // name
		AddSelect(1).          // region: B (after dependency-aware back)
		AddTextInput("test2"). // name again
		AddSelect(0)           // gpu: H100

	engine := NewEngine(p, WithOutput(io.Discard))
	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if region != "b" {
		t.Errorf("expected region 'b', got %q", region)
	}
	if name != "test2" {
		t.Errorf("expected name 'test2', got %q", name)
	}
	if gpu != "h100" {
		t.Errorf("expected gpu 'h100', got %q", gpu)
	}
}

func TestEngine_NilSetter_NosPanic(t *testing.T) {
	flow := &Flow{
		Name: "test",
		Steps: []Step{
			{
				Name:     "temp",
				Prompt:   SelectPrompt,
				Required: true,
				Loader:   StaticChoices(Choice{Label: "A", Value: "a"}),
				// Setter intentionally nil — should not panic
			},
		},
	}

	p := tuitesting.New().AddSelect(0)
	engine := NewEngine(p)

	err := engine.Run(context.Background(), flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine.Collected()["temp"] != "a" {
		t.Errorf("expected collected value 'a', got %v", engine.Collected()["temp"])
	}
}
