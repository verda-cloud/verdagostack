# VerdaGoStack

Shared Go libraries for the Verda platform.

- **Module**: `github.com/verda-cloud/verdagostack`
- **Language**: Go 1.25+

## Installation

```bash
go get github.com/verda-cloud/verdagostack/pkg/log
```

## Packages

| Package | Description |
|---------|-------------|
| [pkg/app](pkg/app/) | CLI application framework built on Cobra, Viper, and pflag |
| [pkg/db](pkg/db/) | Database & cache client constructors — PostgreSQL, CockroachDB, MySQL, Valkey |
| [pkg/errorsx](pkg/errorsx/) | Structured error type with HTTP status codes and metadata (stdlib only) |
| [pkg/errorsx/grpcstatus](pkg/errorsx/grpcstatus/) | gRPC bridge for errorsx — HTTP↔gRPC code mapping |
| [pkg/ginx](pkg/ginx/) | Gin extensions — multi-source binding, typed request handlers, structured error responses |
| [pkg/log](pkg/log/) | Production-grade structured logging backed by zap |
| [pkg/metrics](pkg/metrics/) | Shared Prometheus metrics library — register, update, and serve counters, gauges, histograms, summaries |
| [pkg/middleware/gin](pkg/middleware/gin/) | Gin middleware for OTel trace header injection and structured request logging |
| [pkg/middleware/grpc](pkg/middleware/grpc/) | gRPC interceptor for OTel trace metadata injection and structured request logging |
| [pkg/options](pkg/options/) | Reusable option structs for servers, databases, caches, service discovery, and observability |
| [pkg/otel](pkg/otel/) | OpenTelemetry provider setup, traces + metrics, span error helpers |
| [pkg/server](pkg/server/) | Server interface with HTTP, gRPC, Gin, and Kratos implementations |
| [pkg/util/copier](pkg/util/copier/) | Deep-copy helpers with time.Time ↔ timestamppb.Timestamp converters |
| [pkg/util/ip](pkg/util/ip/) | Local and remote IP address utilities |
| [pkg/util/pagination](pkg/util/pagination/) | Page-based offset calculation |
| [pkg/util/reflect](pkg/util/reflect/) | Struct field inspection, selective field copy, YAML-based deep copy |
| [pkg/util/retry](pkg/util/retry/) | Exponential backoff, polling, and periodic execution (stdlib only) |
| [pkg/util/stringsx](pkg/util/stringsx/) | String slice utilities — Diff, Unique, Contains, Filter, Reverse, base64 |
| [pkg/tui](pkg/tui/) | Interactive TUI framework — Prompter, Status (spinners, progress, tables, pager) |
| [pkg/tui/bubbletea](pkg/tui/bubbletea/) | Bubble Tea v2 backend — select, multiselect, text input, confirm, password, editor, 8 themes |
| [pkg/tui/wizard](pkg/tui/wizard/) | Multi-step wizard engine — flows, steps, Store, MessageBus, ProgressView, HintBarView |
| [pkg/version](pkg/version/) | Build-time version info via `-ldflags` with `--version` flag support |

## Building Apps with verdagostack

AI coding skills are available to help scaffold complete applications following verdagostack conventions. Install them once and your AI assistant (Cursor, Codex, Claude Code) can generate fully-wired services and CLI tools.

### Available skills

| Skill | What it does |
|-------|-------------|
| **verdagostack-web-app** | Scaffold a Gin HTTP service with OTel, options pattern, typed handlers, and observability |
| **verdagostack-cli-app** | Scaffold a Cobra CLI tool with IOStreams, CommandGroups, and Complete/Validate/Run lifecycle |

### Install (Cursor)

From the root of this repo:

```bash
cp -r .ai/skills/verdagostack-web-app ~/.cursor/skills/
cp -r .ai/skills/verdagostack-cli-app ~/.cursor/skills/
```

The skills are stored under `~/.cursor/skills/` and activate automatically in any project.

### Install (Codex / Claude Code)

Copy the skills into your target project's `.ai/skills/` directory:

```bash
cp -r .ai/skills/verdagostack-web-app /path/to/your-app/.ai/skills/
cp -r .ai/skills/verdagostack-cli-app /path/to/your-app/.ai/skills/
```

Then reference them in your project's `AGENTS.md` (or `CLAUDE.md`) skills table.

### Usage

Once installed, just ask your AI assistant in your app project:

- *"Create a new API server called payment-service using verdagostack"* — triggers the web-app skill
- *"Build a CLI tool for managing deployments"* — triggers the CLI-app skill

See [goapp-demo](https://github.com/verda-cloud/goapp-demo) for a complete working example built with these patterns.

## TUI quick start

The `pkg/tui` package provides interactive terminal prompts, spinners, and a multi-step wizard engine built on [Bubble Tea v2](https://github.com/charmbracelet/bubbletea).

```bash
go get github.com/verda-cloud/verdagostack/pkg/tui
```

### Prompts

```go
import (
    "github.com/verda-cloud/verdagostack/pkg/tui"
    _ "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea" // register backend
)

prompter := tui.Default()

// Select with type-to-filter
idx, _ := prompter.Select(ctx, "Pick a color", []string{"Red", "Green", "Blue"})

// MultiSelect
indices, _ := prompter.MultiSelect(ctx, "Toppings", []string{"Cheese", "Pepperoni", "Mushrooms"})

// Text input, confirm, password, editor
name, _ := prompter.TextInput(ctx, "Name", tui.WithDefault("world"))
ok, _ := prompter.Confirm(ctx, "Continue?")
```

### Themes

8 built-in themes (5 dark, 3 light):

| Theme | Background |
|-------|-----------|
| `default` | Dark (ANSI) |
| `dracula` | Dark |
| `catppuccin` | Dark |
| `nord` | Dark |
| `tokyonight` | Dark |
| `github-light` | Light |
| `catppuccin-latte` | Light |
| `solarized-light` | Light |

```go
import "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea"

bubbletea.SetThemeByName("dracula")      // set by name
name := bubbletea.GetThemeName()          // get current
names := bubbletea.ThemeNames()           // list all
```

### Wizard engine

Build multi-step interactive flows with progress bar and contextual hint bar:

```go
import "github.com/verda-cloud/verdagostack/pkg/tui/wizard"

flow := &wizard.Flow{
    Name: "setup",
    Layout: []wizard.ViewDef{
        {ID: "progress", View: wizard.NewProgressView(wizard.WithProgressPercent())},
        {ID: "hints", View: wizard.NewHintBarView()},
    },
    Steps: []wizard.Step{
        {Name: "name", Prompt: wizard.TextInputPrompt, Required: true, ...},
        {Name: "type", Prompt: wizard.SelectPrompt, Loader: wizard.StaticChoices(...), ...},
    },
}

engine := wizard.NewEngine(prompter, status, wizard.WithOutput(os.Stderr))
engine.Run(ctx, flow)
```

Features: back navigation, dependency-aware step loading, choice caching, Store + MessageBus for views.

## Options & Database quick start

The `pkg/options` package provides flag-driven configuration structs that implement `IOptions`:

```go
type IOptions interface {
    Validate() []error
    AddFlags(fs *pflag.FlagSet, fullPrefix string)
}
```

Every option type ships with a `New*Options()` constructor that returns production-ready defaults,
an `AddFlags` method that registers CLI flags under a dot-separated prefix, and a `Validate`
method for startup-time checks. See the [pkg/options README](pkg/options/) for the full catalogue.

### Composing options in your application

```go
import (
    "github.com/verda-cloud/verdagostack/pkg/options"
)

type ServerOptions struct {
    GRPC  *options.GRPCOptions
    DB    *options.PostgreSQLOptions
    Cache *options.ValkeyOptions
}

func NewServerOptions() *ServerOptions {
    return &ServerOptions{
        GRPC:  options.NewGRPCOptions(),
        DB:    options.NewPostgreSQLOptions(),
        Cache: options.NewValkeyOptions(),
    }
}

func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
    o.GRPC.AddFlags(fs, "server.grpc")   // --server.grpc.addr, --server.grpc.timeout, ...
    o.DB.AddFlags(fs, "db.postgresql")    // --db.postgresql.host, --db.postgresql.port, ...
    o.Cache.AddFlags(fs, "cache.valkey")  // --cache.valkey.host, --cache.valkey.tls.cert-path, ...
}
```

### Creating database and cache clients

`pkg/db` constructors accept the option structs from `pkg/options`:

```go
import (
    "github.com/verda-cloud/verdagostack/pkg/db"
    gormadapter "github.com/verda-cloud/verdagostack/pkg/log/adapter/gorm"
)

// PostgreSQL / CockroachDB / MySQL — all return *gorm.DB
gormDB, err := db.NewPostgreSQL(opts.DB, gormadapter.New(logger))
crDB,   err := db.NewCockroachDB(crdbOpts, nil) // nil → default GORM logger

// Valkey / Redis — returns valkey.Client
valkeyClient, err := db.NewValkey(opts.Cache)
```
