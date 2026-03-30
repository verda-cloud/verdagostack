# pkg/log

Production-grade structured logging for verdastack, backed by [zap](https://github.com/uber-go/zap).

## Features

- **Dual API**: printf-style (`Infof`) and structured key-value (`Infow`) at every level
- **Six severity levels**: Debug, Info, Warn, Error, Panic, Fatal
- **Context-aware logging**: `W(ctx)` extracts request-scoped fields (request ID, trace ID)
- **Child loggers**: `With("component", "auth")` for pre-set fields
- **CLI integration**: `--log.level`, `--log.format`, `--log.output-paths` via pflag
- **Zap backend**: stacktrace on panic, multi-output paths, JSON/console encoding, caller info
- **Framework adapters**: GORM, Kratos, Gin (separate sub-packages, not baked into the interface)
- **Testable**: no-op `empty.Logger` for unit tests

## Quick Start

```go
import "github.com/verda-cloud/verdagostack/pkg/log"

// Use the global logger immediately — zero configuration required
log.Infow("server starting", "port", 8080)
log.Infof("listening on :%d", 8080)
```

## Initialization

### Basic

```go
opts := log.NewOptions()
opts.Level = "debug"
opts.Format = "json"
log.Init(opts)
defer log.Sync()
```

### With pflag CLI flags

```go
import "github.com/spf13/pflag"

opts := log.NewOptions()
opts.AddFlags(pflag.CommandLine)
pflag.Parse()
log.Init(opts)
defer log.Sync()
```

This registers:

| Flag | Default | Description |
|------|---------|-------------|
| `--log.level` | `info` | Minimum level: debug, info, warn, error, dpanic, panic, fatal |
| `--log.format` | `console` | Output encoding: console or json |
| `--log.output-paths` | `stdout` | Comma-separated output destinations |
| `--log.disable-caller` | `false` | Omit caller file:line from output |
| `--log.disable-stacktrace` | `false` | Omit stack traces at panic level |
| `--log.enable-color` | `false` | ANSI color output (console format only) |

### Multi-output paths

```go
opts := &log.Options{
    Level:       "info",
    Format:      "json",
    OutputPaths: []string{"stdout", "/var/log/myapp.log"},
}
log.Init(opts)
```

## Logging

### Printf-style (`*f`)

```go
log.Debugf("processing item %d of %d", i, total)
log.Infof("user %s logged in from %s", username, ip)
log.Errorf("failed to connect: %v", err)
```

### Structured (`*w`)

```go
log.Debugw("processing item", "current", i, "total", total)
log.Infow("user logged in", "username", username, "ip", ip)
log.Errorw("connection failed", "err", err, "host", host, "port", port)
```

### Context-aware (`W`)

Extract request-scoped fields automatically:

```go
log.Init(opts, log.WithContextExtractor(log.ContextExtractors{
    "request.id": func(ctx context.Context) string {
        return middleware.RequestIDFromContext(ctx)
    },
    "trace.id": func(ctx context.Context) string {
        return tracing.TraceIDFromContext(ctx)
    },
}))

// Later, in a request handler:
log.W(ctx).Infow("handling request", "method", "GET", "path", "/api/users")
// Output includes: {"request.id": "abc-123", "trace.id": "xyz-789", ...}
```

### Child loggers (`With`)

```go
authLog := log.With("component", "auth", "version", "v2")
authLog.Infow("token issued", "user", "alice")
// Output: {"component": "auth", "version": "v2", "user": "alice", ...}
```

## Dependency Injection

Pass the `Logger` interface to constructors instead of using the global logger directly:

```go
type OrderService struct {
    log log.Logger
}

func NewOrderService(logger log.Logger) *OrderService {
    return &OrderService{log: logger.With("service", "orders")}
}

func (s *OrderService) Create(ctx context.Context, order Order) error {
    s.log.W(ctx).Infow("creating order", "order_id", order.ID)
    // ...
}
```

Wire it up in main:

```go
svc := NewOrderService(log.Default())
```

## Framework Adapters

### GORM

```go
import logadaptergorm "github.com/verda-cloud/verdagostack/pkg/log/adapter/gorm"

gormLogger := logadaptergorm.New(log.Default())
// Optional: custom slow-query threshold (default 200ms)
gormLogger = gormLogger.WithSlowThreshold(500 * time.Millisecond)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: gormLogger,
})
```

Produces structured output for SQL queries with `elapsed_ms`, `rows`, `sql`, and warns on slow queries.

### Kratos

```go
import logadapterkratos "github.com/verda-cloud/verdagostack/pkg/log/adapter/kratos"

kratosLogger := logadapterkratos.New(log.Default())
krtlog.SetLogger(kratosLogger)
```

### Gin

```go
import logadaptergin "github.com/verda-cloud/verdagostack/pkg/log/adapter/gin"

// Option 1: Redirect Gin's internal debug output through your logger
gin.DefaultWriter = logadaptergin.Writer(log.Default(), logadaptergin.LevelInfo)
gin.DefaultErrorWriter = logadaptergin.Writer(log.Default(), logadaptergin.LevelError)

// Option 2: Structured request logging middleware (replaces gin.Logger())
router := gin.New()
router.Use(gin.Recovery())
router.Use(logadaptergin.Middleware(log.Default(), &logadaptergin.MiddlewareConfig{
    SkipPaths:     []string{"/health", "/readyz"},
    SlowThreshold: 200 * time.Millisecond,
}))
```

The middleware logs every request with structured fields:

```json
{"level":"info","message":"HTTP request","method":"GET","path":"/api/users","status":200,"latency_ms":1.23,"client_ip":"10.0.0.1","user_agent":"curl/8.0"}
```

- 2xx -> info, 4xx -> warn, 5xx -> error
- Slow requests (above threshold) escalate to warn with `"slow": true`
- Skipped paths produce no log output
- Context fields (request ID, trace ID) included via `W(c.Request.Context())`

## Testing

Use the no-op logger to silence output in tests:

```go
import "github.com/verda-cloud/verdagostack/pkg/log/empty"

func TestOrderService_Create(t *testing.T) {
    svc := NewOrderService(empty.NewLogger())
    // ...
}
```

## Package Layout

```
pkg/log/
  doc.go                        Package documentation
  log.go                        Logger interface, zap implementation, global functions
  options.go                    Options struct, NewOptions(), pflag integration
  context.go                    ContextExtractors, W(ctx)
  adapter/
    gorm/gorm.go                GORM logger adapter
    kratos/kratos.go            Kratos logger adapter
    gin/gin.go                  Gin io.Writer bridge + request middleware
  empty/empty.go                No-op Logger for testing
  example/main.go               Runnable demo of all features
```

## Running the Example

```bash
# Default (console, info level)
go run ./pkg/log/example

# JSON output at debug level
go run ./pkg/log/example --log.level=debug --log.format=json

# Color output
go run ./pkg/log/example --log.enable-color
```
