// Package db provides constructors for creating database and cache clients
// used across the verdagostack project.
//
// Supported backends:
//   - PostgreSQL (via GORM + pgx)
//   - CockroachDB (PostgreSQL-compatible, via GORM + pgx)
//   - MySQL (via GORM + mysql driver)
//   - Valkey / Redis (via valkey-go)
//
// Connection options for each backend are defined in [pkg/options] (e.g.,
// [options.PostgreSQLOptions], [options.ValkeyOptions]).
//
// # SQL backends (PostgreSQL, CockroachDB, MySQL)
//
//	opts := options.NewPostgreSQLOptions()
//	gormDB, err := db.NewPostgreSQL(opts, nil) // nil logger → use default GORM logger
//
// # Valkey / Redis
//
//	opts := options.NewValkeyOptions()
//	client, err := db.NewValkey(opts)
package db // import "github.com/verda-cloud/verdagostack/pkg/db"
