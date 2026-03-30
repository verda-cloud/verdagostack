# pkg/options

Reusable, flag-driven configuration types for verdastack services.

Every option struct implements the `IOptions` interface:

```go
type IOptions interface {
    Validate() []error
    AddFlags(fs *pflag.FlagSet, fullPrefix string)
}
```

## Available options

### Server & Networking

| Type | File | Description |
|------|------|-------------|
| `GRPCOptions` | `grpc_options.go` | gRPC server network, address, and timeout |
| `InsecureServingOptions` | `insecure_serving.go` | Plain HTTP server address and timeout |
| `SecureServingOptions` | `secure_serving.go` | HTTPS/TLS server address, timeout, and certs |
| `TLSOptions` | `tls_options.go` | TLS certificates (file path, PEM, or Base64) |

### Databases & Caching

| Type | File | Description |
|------|------|-------------|
| `PostgreSQLOptions` | `postgresql_options.go` | PostgreSQL host, port, credentials, pool, SSL mode |
| `CockroachDBOptions` | `cockroachdb_options.go` | CockroachDB (PostgreSQL-compatible) with timezone support |
| `MySQLOptions` | `mysql_options.go` | MySQL host, port, credentials, connection pool |
| `ValkeyOptions` | `valkey_options.go` | Valkey/Redis host, port, TLS, Sentinel, TTL |

### Service Discovery & Observability

| Type | File | Description |
|------|------|-------------|
| `ConsulOptions` | `consul_options.go` | Consul address and scheme |
| `EtcdOptions` | `etcd_options.go` | etcd endpoints, credentials, TLS, dial timeout |
| `PolarisOptions` | `polaris_options.go` | Polaris service discovery and heartbeat |
| `ObservabilityOptions` | `observability_options.go` | Health checks, Prometheus metrics, pprof |
| `HealthOptions` | `health_options.go` | Legacy health check HTTP server |

## Usage

### Registering flags

Use `AddFlags` with a dot-separated prefix. The prefix is prepended to every flag name:

```go
pgOpts := options.NewPostgreSQLOptions()
pgOpts.AddFlags(fs, "store.postgresql")
// Registers: --store.postgresql.host, --store.postgresql.port, ...

valkeyOpts := options.NewValkeyOptions()
valkeyOpts.AddFlags(fs, "cache.valkey")
// Registers: --cache.valkey.host, --cache.valkey.tls.cert-path, ...
```

### Composing in application options

Embed option structs in your application's top-level options struct and use the `Join` helper to build prefixes:

```go
type AppOptions struct {
    GRPC    *options.GRPCOptions
    DB      *options.PostgreSQLOptions
    Cache   *options.ValkeyOptions
}

func (o *AppOptions) AddFlags(fs *pflag.FlagSet) {
    o.GRPC.AddFlags(fs, "server.grpc")
    o.DB.AddFlags(fs, "db.postgresql")
    o.Cache.AddFlags(fs, "cache.valkey")
}

func (o *AppOptions) Validate() []error {
    var errs []error
    errs = append(errs, o.GRPC.Validate()...)
    errs = append(errs, o.DB.Validate()...)
    errs = append(errs, o.Cache.Validate()...)
    return errs
}
```

### Creating database clients

Options provide `DSN()` methods for SQL backends. Pass the options to constructors in `pkg/db`:

```go
import "github.com/verda-cloud/verdagostack/pkg/db"

pgOpts := options.NewPostgreSQLOptions()
// ... configure or parse flags ...

gormDB, err := db.NewPostgreSQL(pgOpts, nil) // nil = default GORM logger
```

### Valkey with Sentinel and mTLS

```go
opts := options.NewValkeyOptions()
opts.Sentinel.MasterName = "mymaster"
opts.Sentinel.Addrs = []string{"sentinel1:26379", "sentinel2:26379"}
opts.TLS.CertPath = "/certs/client.crt"
opts.TLS.KeyPath  = "/certs/client.key"
opts.TLS.CACertPath = "/certs/ca.crt"

client, err := db.NewValkey(opts)
```

## Helpers

- `Join(prefixes ...string) string` — joins prefixes with `.` separators and adds a trailing `.`.
- `ValidateAddress(addr string) error` — validates `host:port` format.
- `GetLocalIP() string` — returns the first non-loopback IPv4 address.
