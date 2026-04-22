// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
