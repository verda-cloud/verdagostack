// Package wizard provides a TUI wizard engine for interactive CLI flows.
//
// # Overview
//
// A wizard Flow is an ordered list of Steps executed sequentially. Each step
// collects one value from the user via a prompt (select, text input, confirm,
// password, or multi-select). The engine handles skip logic, back navigation,
// validation, default values, and dependency-driven choice loading.
//
// The engine uses a composable region/actor layout. Each visual region is an
// independent actor with its own mailbox — it receives typed messages and
// renders independently. The engine acts as a message bus, broadcasting
// lifecycle events and routing inter-region messages.
//
// # Quick Start
//
//	var region string
//
//	flow := &wizard.Flow{
//	    Name: "create-vm",
//	    Steps: []wizard.Step{
//	        {
//	            Name:     "region",
//	            Prompt:   wizard.SelectPrompt,
//	            Required: true,
//	            Loader:   wizard.StaticChoices(
//	                wizard.Choice{Label: "Finland", Value: "FIN-01"},
//	                wizard.Choice{Label: "Sweden",  Value: "SWE-01"},
//	            ),
//	            Setter: func(v any) { region = v.(string) },
//	        },
//	    },
//	}
//
//	engine := wizard.NewEngine(prompter, status)
//	if err := engine.Run(ctx, flow); err != nil {
//	    return err
//	}
//	fmt.Println("Selected:", region)
//
// # Engine Constructor
//
// NewEngine takes a [tui.Prompter] for collecting input and an optional
// [tui.Status] for spinners/progress during async operations:
//
//	// Minimal — no spinners, default progress bar.
//	engine := wizard.NewEngine(prompter, nil)
//
//	// With status support and custom output.
//	engine := wizard.NewEngine(prompter, status, wizard.WithOutput(os.Stderr))
//
// The bubbletea.Prompter implements both tui.Prompter and tui.Status, so you
// can pass the same object for both:
//
//	p := bubbletea.New()
//	engine := wizard.NewEngine(p, p)
//
// # Steps
//
// Each Step defines what to collect and how:
//
//	wizard.Step{
//	    Name:        "instance-type",          // unique key in collected map
//	    Description: "Select instance type",   // shown in the prompt
//	    Prompt:      wizard.SelectPrompt,       // widget type
//	    Required:    true,                      // enforce non-empty
//	    DependsOn:   []string{"region"},        // invalidate when region changes
//
//	    // Loader fetches choices. Receives Status for spinners and Store for shared data.
//	    Loader: func(ctx context.Context, p tui.Prompter, s tui.Status, store *wizard.Store) ([]wizard.Choice, error) {
//	        c := store.Collected()
//	        region := c["region"].(string)
//
//	        sp, _ := s.Spinner(ctx, "Loading instance types...")
//	        types, err := api.ListInstanceTypes(ctx, region)
//	        sp.Stop("Loaded")
//
//	        return buildChoices(types), err
//	    },
//
//	    // Default value based on previously collected values.
//	    Default: func(c map[string]any) any { return "standard-2vcpu" },
//
//	    // Validate before accepting.
//	    Validate: func(v any) error { return nil },
//
//	    // Setter writes to your options struct.
//	    Setter:   func(v any) { opts.InstanceType = v.(string) },
//	    Resetter: func()      { opts.InstanceType = "" },
//
//	    // Skip if already provided via CLI flag.
//	    IsSet: func() bool { return opts.InstanceType != "" },
//	    Value: func() any  { return opts.InstanceType },
//
//	    // Conditional skip based on earlier answers.
//	    ShouldSkip: func(c map[string]any) bool { return c["env"] == "dev" },
//	}
//
// # Prompt Types
//
//   - [SelectPrompt] — single choice from a list (arrow keys + type-to-filter)
//   - [MultiSelectPrompt] — multiple choices (space to toggle)
//   - [TextInputPrompt] — free-form text
//   - [ConfirmPrompt] — yes/no
//   - [PasswordPrompt] — masked input
//
// # Loaders
//
// Loaders fetch choices for select/multi-select steps. They receive four arguments:
//
//   - ctx — context for cancellation
//   - prompter — for running sub-prompts (e.g., "Create SSH key?")
//   - status — for showing spinners/progress during API calls (may be nil)
//   - store — shared data layer; use store.Collected() for step values
//
// For fixed choices, use the StaticChoices helper:
//
//	Loader: wizard.StaticChoices(
//	    wizard.Choice{Label: "Small",  Value: "small"},
//	    wizard.Choice{Label: "Large",  Value: "large"},
//	)
//
// For dynamic choices loaded from an API:
//
//	Loader: func(ctx context.Context, _ tui.Prompter, s tui.Status, store *wizard.Store) ([]wizard.Choice, error) {
//	    c := store.Collected()
//	    region := c["region"].(string)
//
//	    // Show spinner while loading.
//	    sp, _ := s.Spinner(ctx, "Fetching sizes...")
//	    sizes, err := api.ListSizes(ctx, region)
//	    sp.Stop("Done")
//
//	    // Write to store — other regions can react to this.
//	    store.Set("cheapest", findCheapest(sizes))
//
//	    return toChoices(sizes), err
//	}
//
// # Dependencies
//
// When a step's DependsOn list includes another step name, the engine
// invalidates cached loader results when that dependency changes. This
// ensures choices are re-fetched when the user goes back and changes
// an earlier answer.
//
// # Back Navigation
//
// The engine automatically adds a "← Back" option to select/multi-select
// prompts when there is an editable prior step. Pressing Esc on any prompt
// also navigates back. The engine clears and resets all steps between the
// current position and the target.
//
// # Store
//
// The Store is a shared data layer accessible to loaders and regions:
//
//	store.Collected()              // snapshot of step name → value
//	store.Get("cost")              // read arbitrary data
//	store.Set("cost", 3.50)        // write arbitrary data
//
// Collected values are managed by the engine (set on step completion,
// cleared on back navigation). Arbitrary data is managed by your code —
// loaders can write values that custom regions display.
//
// After the flow completes, read results via:
//
//	engine.Collected()             // same as store.Collected()
//	engine.Store().Get("cost")     // arbitrary data written during the flow
//
// # Layout and Regions
//
// The engine renders the terminal as a vertical stack of regions. Each region
// is an independent actor that receives messages and renders output.
//
// Default layout (when Flow.Layout is nil):
//
//	┌──────────────────────────────────────┐
//	│ ━━━━━━━━━━━━━━░░░░░░  Step 3 of 10  │  ← ProgressRegion
//	│ ? Select instance type               │  ← Prompt (engine's sequential loop)
//	│   > Standard-2vCPU-8GB               │
//	│     Standard-4vCPU-16GB              │
//	└──────────────────────────────────────┘
//
// Custom layout with additional regions:
//
//	flow := &wizard.Flow{
//	    Layout: []wizard.RegionDef{
//	        {ID: "progress", Region: wizard.NewProgressRegion()},
//	        // Custom cost region that updates when collected values change.
//	        {ID: "cost", Region: &CostRegion{}},
//	    },
//	    Steps: []wizard.Step{...},
//	}
//
// # Region Interface
//
// A region implements two methods:
//
//	type Region interface {
//	    Update(msg any) (render string, publish []any)
//	    Subscribe() []reflect.Type
//	}
//
// Update receives a message and returns the new display string plus optional
// messages to publish to other regions. Subscribe returns the message types
// this region listens to (nil = engine broadcasts only).
//
// # Built-in Messages
//
// The engine broadcasts these messages:
//
//   - [StepChangedMsg] — when the engine moves to a new step (contains
//     Current, Total, StepName, Collected)
//   - [CollectedChangedMsg] — when a step completes (contains Key, Value,
//     Collected)
//   - [StoreChangedMsg] — when a store value is set (contains Key, Value)
//
// # Custom Region Example
//
// A region that shows estimated cost, updating when collected values change:
//
//	type CostRegion struct {
//	    last string
//	}
//
//	func (r *CostRegion) Update(msg any) (string, []any) {
//	    if m, ok := msg.(wizard.CollectedChangedMsg); ok {
//	        if price, ok := calculatePrice(m.Collected); ok {
//	            r.last = fmt.Sprintf("  Estimated cost: $%.2f/hr\n", price)
//	        }
//	    }
//	    return r.last, nil
//	}
//
//	func (r *CostRegion) Subscribe() []reflect.Type {
//	    return nil // receive all engine broadcasts
//	}
//
// # Inter-Region Messaging
//
// Regions can communicate by publishing typed messages. The engine routes
// published messages to regions that subscribe to those types:
//
//	// Region A publishes DataReadyMsg.
//	func (r *LoaderRegion) Update(msg any) (string, []any) {
//	    return r.last, []any{DataReadyMsg{Items: 42}}
//	}
//
//	// Region B subscribes to DataReadyMsg.
//	func (r *SummaryRegion) Subscribe() []reflect.Type {
//	    return []reflect.Type{reflect.TypeFor[DataReadyMsg]()}
//	}
//
// Messages chain: if Region B publishes messages in response, those are
// delivered to their subscribers in the same cycle.
//
// # Step Lifecycle
//
// Each step goes through one of these states:
//
//   - Pending — not yet visited
//   - Fixed — IsSet() returned true, value provided via flag/config (skipped)
//   - Skipped — ShouldSkip() returned true (skipped)
//   - AutoSkipped — loader returned empty choices for optional step (skipped)
//   - Completed — user answered or default applied
//
// Only Completed and Fixed steps appear in Collected(). The progress bar
// counts only visible steps (not Fixed, Skipped, or AutoSkipped).
package wizard
