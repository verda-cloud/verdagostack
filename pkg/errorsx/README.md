# pkg/errorsx

Structured error type for APIs carrying HTTP status codes, business reason strings, messages, and metadata.

## Design

The package is split into two layers to keep dependencies minimal:

| Package | Dependencies | Purpose |
|---------|-------------|---------|
| `pkg/errorsx` | **stdlib only** | Core `ErrorX` type, constructors, matching |
| `pkg/errorsx/grpcstatus` | gRPC, protobuf | HTTP↔gRPC code mapping, `ToGRPCStatus`, gRPC-aware `FromError` |

HTTP/CLI code imports only `errorsx`. gRPC services additionally import `grpcstatus`.

## Quick Start

```go
import "github.com/verda-cloud/verdagostack/pkg/errorsx"

// Use predefined sentinel errors
err := errorsx.ErrNotFound.WithMessage("user %d not found", userID)

// Create custom errors
err := errorsx.New(http.StatusConflict, "DuplicateEmail", "email %s already exists", email)

// Add metadata
err = err.KV("field", "email", "value", email)

// Add request ID for tracing
err = err.WithRequestID(requestID)
```

## Sentinel Errors

Predefined errors in `code.go`:

| Variable | HTTP Code | Reason |
|----------|-----------|--------|
| `ErrInternal` | 500 | InternalError |
| `ErrNotFound` | 404 | NotFound |
| `ErrBind` | 400 | BindError |
| `ErrInvalidArgument` | 400 | InvalidArgument |
| `ErrUnauthenticated` | 401 | Unauthenticated |
| `ErrPermissionDenied` | 403 | PermissionDenied |
| `ErrOperationFailed` | 409 | OperationFailed |

All `WithMessage`, `WithMetadata`, and `KV` methods return copies — sentinel errors are never mutated.

## Error Matching

`ErrorX` implements `Is()` matching by `Code` + `Reason`, so `errors.Is` works naturally:

```go
err := errorsx.ErrNotFound.WithMessage("user 42 not found")
errors.Is(err, errorsx.ErrNotFound) // true — same Code and Reason
```

## Converting Errors

```go
// Convert any error to *ErrorX (non-ErrorX errors become 500 Internal)
errx := errorsx.FromError(err)

// Extract HTTP code from any error
code := errorsx.Code(err) // 404, 500, etc.
```

## gRPC Integration

For gRPC services, import the `grpcstatus` sub-package:

```go
import "github.com/verda-cloud/verdagostack/pkg/errorsx/grpcstatus"

// Convert ErrorX to gRPC status (with ErrorInfo details)
gs := grpcstatus.ToGRPCStatus(err)

// Convert any error (including gRPC status errors) to *ErrorX
errx := grpcstatus.FromError(err)
```
