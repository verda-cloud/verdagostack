# Verdastack AI Agent Guide

This file tells AI coding assistants (Cursor, Codex, Claude Code, Copilot, etc.) how to work with this repository.

## Project

- **Module**: `github.com/verda-cloud/verdagostack`
- **Language**: Go (1.25+)
- **Build**: `go build ./...`
- **Test**: `go test ./... -count=1`
- **Vet**: `go vet ./...`

## Skills

Reusable step-by-step guides for common tasks live in `.ai/skills/`. Read the relevant skill before starting work.

| Skill | Path | When to use |
|-------|------|-------------|
| Log adapter | [.ai/skills/verdagostack-log-adapter/SKILL.md](.ai/skills/verdagostack-log-adapter/SKILL.md) | Creating a new framework adapter for `pkg/log` (e.g., integrating Redis, gRPC, or any third-party library with verdagostack logging) |

## Conventions

### Logging

All logging goes through `pkg/log`. See [pkg/log/README.md](pkg/log/README.md) for full usage.

- Use `log.Infow` / `log.Errorf` (structured `*w` preferred over printf `*f` for production code)
- Pass `log.Logger` interface via constructor for dependency injection
- Use `log.W(ctx)` in request handlers to include request-scoped fields
- Use `empty.NewLogger()` in tests
- Framework integrations live in `pkg/log/adapter/<framework>/` as separate packages
- Never embed framework logger interfaces into `log.Logger`

### Code style

- Run `go vet ./...` before committing
- New Go source files need the standard Apache 2.0 file header: run `make license` to add it; CI runs `make license-check`
- Tests required for all new packages
- Structured key-value logging preferred over printf-style in application code
