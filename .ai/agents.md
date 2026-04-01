# AI Agents Guide

## Wizard Engine Development

When working on the wizard engine (`pkg/tui/wizard/`):

### Key Files
- `wizard.go` — Types: Step, Flow, Choice, LoaderFunc, ViewDef
- `engine.go` — Engine: Run loop, handlePrompt, transitions, navigation
- `view.go` — View interface, message types (StepChangedMsg, CollectedChangedMsg, StoreChangedMsg)
- `store.go` — Store shared data layer
- `bus.go` — MessageBus: Broadcast, Publish, RenderChanged
- `view_progress.go` — Built-in ProgressView
- `doc.go` — Comprehensive documentation

### Architecture
- Views are synchronous actors — no queues, no goroutines
- Message bus routes by Go struct type (reflect.TypeOf)
- Engine broadcasts lifecycle events to all views
- Views publish inter-view messages routed to subscribers only
- Only changed view output is printed (deduplication via RenderChanged)

### Breaking Change Protocol
When changing LoaderFunc signature or View interface:
1. Update the interface/type in wizard.go or view.go
2. Update StaticChoices helper
3. Update engine.go (loadChoices, Run, handlePrompt)
4. Update ALL tests in engine_test.go (many loaders to fix)
5. Update examples in pkg/tui/examples/wizard/ and wizard-views/
6. Update doc.go examples
7. Run `go build ./...` and `go test ./...`

### Testing Pattern
Tests use `pkg/tui/testing.Prompter` mock:
```go
p := tuitesting.New().AddSelect(0).AddTextInput("value")
engine := wizard.NewEngine(p, nil, wizard.WithOutput(&buf))
err := engine.Run(ctx, flow)
```

## Bubbletea Components

When working on TUI components (`pkg/tui/bubbletea/`):

### v2 API Notes
- `View()` returns `tea.View` (not string) — use `tea.NewView(s)`
- Key handling uses `tea.KeyPressMsg` with `msg.Code` and `msg.Text`
- lipgloss v2 has no renderer — styles are pure values
- Import paths: `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`

### Adding a New Prompt Type
1. Create `pkg/tui/bubbletea/<name>.go` with bubbletea Model
2. Add method to `tui.Prompter` interface in `pkg/tui/tui.go`
3. Add options to `pkg/tui/tui.go` (Config struct + Option type + WithX helpers)
4. Add `ResolveConfig` to `pkg/tui/options.go`
5. Implement on `bubbletea.Prompter`
6. Add test mock to `pkg/tui/testing/prompter.go`
7. Add example to `pkg/tui/examples/bubbletea/<name>/`

### Adding a New Status Component
1. Create `pkg/tui/bubbletea/<name>.go`
2. Add method to `tui.Status` interface in `pkg/tui/status.go`
3. Add options to `pkg/tui/status_options.go`
4. Implement on `bubbletea.Prompter` (it implements both interfaces)
5. Add test mock to `pkg/tui/testing/prompter.go`
