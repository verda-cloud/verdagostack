# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.1.0] - 2025-03-23

Initial release of the Verdastack shared Go library.

### Added

- **pkg/app** — CLI application framework built on Cobra, Viper, and pflag.
- **pkg/server** — HTTP and gRPC server lifecycle management.
- **pkg/options** — Configuration binding from flags, environment variables, and config files.
- **pkg/otel** — OpenTelemetry tracing and metrics provider with OTLP export.
- **pkg/log** — Structured logging with zap backend and adapter support.
- **pkg/ginx** — Gin middleware: recovery, request ID, logging, metrics.
- **pkg/db** — Database and cache option types for PostgreSQL, CockroachDB, MySQL, Valkey.
- **pkg/errorsx** — Structured error type with HTTP status codes and metadata.
- **pkg/version** — Build-time version info injected via `-ldflags`.
- **pkg/util** — Helpers: reflect copy, ID generation, string utilities.

[Unreleased]: https://github.com/verda-cloud/verdagostack/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/verda-cloud/verdagostack/-/tags/v0.1.0
