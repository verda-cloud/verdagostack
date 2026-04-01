# VerdaGoStack

Shared Go library for the Verda cloud platform. Requires Go 1.25+.

## Project Structure

```
pkg/
  app/          - Cobra CLI framework
  db/           - PostgreSQL/CockroachDB/MySQL/Valkey
  errorsx/      - Error handling + gRPC status mapping
  ginx/         - Typed Gin handlers
  log/          - Zap-based structured logging
  metrics/      - Prometheus metrics
  middleware/    - Gin and gRPC middleware
  otel/         - OpenTelemetry tracing
  options/      - Functional options pattern
  server/       - HTTP/gRPC server
  tui/          - Terminal UI framework
    bubbletea/  - Bubbletea prompt implementations (v2)
    wizard/     - Wizard engine with View actor model
    testing/    - Test doubles for Prompter/Status
  util/         - Copier, IP, pagination, retry, strings
  version/      - Build version info
```

## TUI Architecture

The TUI layer has two interfaces:

- **`tui.Prompter`** — collects user input (select, text, confirm, password, multi-select, editor)
- **`tui.Status`** — displays output (spinner, progress, table, pager)

The `bubbletea` package implements both using charmbracelet v2:
- `charm.land/bubbletea/v2` (v2.0.2)
- `charm.land/bubbles/v2` (v2.1.0)
- `charm.land/lipgloss/v2` (v2.0.2)

## Wizard Engine

The wizard engine (`pkg/tui/wizard`) sequences TUI prompts into step-by-step workflows with a pub/sub message bus for display views.

Key concepts:
- **Steps** — sequential prompts that collect values (select, text, confirm)
- **Views** — display actors that render based on messages (progress bar, cost display)
- **Store** — shared data layer for collected values + arbitrary data
- **MessageBus** — broadcasts engine events, routes inter-view messages by Go struct type

```go
engine := wizard.NewEngine(prompter, status)
engine.Run(ctx, flow)
```

See `pkg/tui/wizard/doc.go` for comprehensive documentation.
See `pkg/tui/examples/wizard-views/` for View pub/sub example.

## Development

```bash
go build ./...          # build
go test ./...           # test
make lint               # golangci-lint
```

## Conventions

- Functional options pattern for configuration (`WithFoo()`)
- Pre-commit hooks enforce: gofmt, goimports, go mod tidy, tests, golangci-lint
- Commit messages follow conventional commits: `feat(scope):`, `fix(scope):`, `refactor(scope):`
- Breaking changes documented in commit body
