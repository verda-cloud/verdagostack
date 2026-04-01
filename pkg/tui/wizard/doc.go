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
// # Layout and Regions (Actor Model)
//
// The engine uses a composable actor model for rendering. The terminal is a
// vertical stack of regions, each an independent actor with its own mailbox.
// The engine acts as a message bus: it broadcasts lifecycle events and routes
// inter-region messages.
//
// Default layout (when Flow.Layout is nil):
//
//	┌─────────────────────────────────────────────┐
//	│ ████████████████░░░░░░░░░░  Step 3 of 10   │  ← ProgressRegion (gradient bar)
//	│ ? Select instance type                      │  ← Prompt (engine's sequential loop)
//	│   > Standard-2vCPU-8GB                      │
//	│     Standard-4vCPU-16GB                     │
//	└─────────────────────────────────────────────┘
//
// # How the Actor Model Works
//
// Each region is an actor. It receives messages, updates its internal state,
// and returns a render string. The engine only prints a region's output when
// it actually changes, so redundant renders are never written to the terminal.
//
// The message flow for one step:
//
//	Engine                         Regions
//	──────                         ───────
//	1. Broadcast StepChangedMsg ──→ ProgressRegion: renders "Step 3 of 10"  (changed → print)
//	                            ──→ CostRegion: no change                   (skip)
//	2. Run prompt (blocks for user input)
//	3. User picks a value
//	4. Broadcast CollectedChanged ─→ ProgressRegion: same output            (unchanged → skip)
//	                              ─→ CostRegion: recalculates "$3.20/hr"    (changed → print)
//	5. Advance to next step
//
// Key design properties:
//
//   - Regions are decoupled: a CostRegion doesn't know about ProgressRegion.
//     They communicate through typed messages, not direct calls.
//   - Only changed output is printed: the engine tracks each region's last
//     printed output and skips unchanged regions. This prevents duplicate
//     renders (e.g., ProgressRegion re-printing the same bar after a value
//     change that doesn't affect it).
//   - Regions can publish messages to each other. The engine routes published
//     messages to subscribers, enabling region-to-region communication.
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
// Update receives a message and returns:
//   - render: the new display string (return the same string if nothing changed)
//   - publish: optional messages to broadcast to other regions (nil if none)
//
// Subscribe returns the message types this region listens to. Return nil to
// receive all engine broadcasts. Return specific types to also receive
// inter-region messages of those types.
//
// # Built-in Messages
//
// The engine broadcasts these messages at specific points in the flow:
//
//   - [StepChangedMsg] — broadcast BEFORE each prompt. Contains Current
//     (1-based step position), Total (number of steps), StepName, and a
//     snapshot of Collected values. Use this for progress indicators.
//   - [CollectedChangedMsg] — broadcast AFTER a step completes. Contains
//     Key (step name), Value (selected value), and the updated Collected
//     snapshot. Use this for reactive displays (cost, summary, etc.).
//   - [StoreChangedMsg] — broadcast when a loader or region calls
//     store.Set(). Contains Key and Value. Use this for displays driven
//     by arbitrary data written during loading.
//
// # Progress Bar
//
// The built-in [ProgressRegion] uses the charmbracelet/bubbles progress
// component for gradient-colored rendering (same style as the bubbletea
// static progress example). It responds to [StepChangedMsg] and shows
// "Step X of Y" by default. Hidden for single-step flows.
//
// Default (bubbles default gradient #5A56E0 → #EE6FF8, step label):
//
//	wizard.NewProgressRegion()
//
// Custom gradient to match your theme:
//
//	wizard.NewProgressRegion(
//	    wizard.WithProgressGradient("#bd93f9", "#ff79c6"),
//	)
//
// Percentage mode (animated progress example style):
//
//	wizard.NewProgressRegion(wizard.WithProgressPercent())
//
// Solid fill, custom width:
//
//	wizard.NewProgressRegion(
//	    wizard.WithProgressSolidFill("#50fa7b"),
//	    wizard.WithProgressWidth(30),
//	)
//
// # Custom Layout
//
// Custom layout with additional regions:
//
//	flow := &wizard.Flow{
//	    Layout: []wizard.RegionDef{
//	        {ID: "progress", Region: wizard.NewProgressRegion(
//	            wizard.WithProgressGradient("#bd93f9", "#ff79c6"),
//	        )},
//	        {ID: "cost", Region: &CostRegion{}},
//	    },
//	    Steps: []wizard.Step{...},
//	}
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
//	    return r.last, nil  // unchanged output → engine skips printing
//	}
//
//	func (r *CostRegion) Subscribe() []reflect.Type {
//	    return nil // receive all engine broadcasts
//	}
//
// # Inter-Region Messaging
//
// Regions can publish messages that other regions subscribe to. The engine
// routes published messages to subscribers only (not broadcast to all):
//
//	// Region A publishes DataReadyMsg when it receives StepChangedMsg.
//	func (r *LoaderRegion) Update(msg any) (string, []any) {
//	    if _, ok := msg.(wizard.StepChangedMsg); ok {
//	        data := fetchData()
//	        return r.render(data), []any{DataReadyMsg{Items: len(data)}}
//	    }
//	    return r.last, nil
//	}
//
//	// Region B subscribes to DataReadyMsg.
//	func (r *SummaryRegion) Subscribe() []reflect.Type {
//	    return []reflect.Type{reflect.TypeFor[DataReadyMsg]()}
//	}
//
// Messages chain: if Region B publishes messages in response to DataReadyMsg,
// those are delivered to their subscribers in the same cycle.
//
// # Pager
//
// For displaying long content (lists, logs, details), use [tui.Status.Pager]:
//
//	status.Pager(ctx, content)
//	status.Pager(ctx, content, tui.WithPagerTitle("Volumes in trash"))
//
// The pager auto-detects: if content fits the terminal, it prints directly.
// If it overflows, it shows an interactive scrollable viewport (alt-screen)
// with arrow keys/j/k/pgup/pgdn navigation and q/esc to exit.
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
// uses absolute step position (Step 3 of 12), so the total is stable and
// the bar always advances — even when steps are skipped.
package wizard
