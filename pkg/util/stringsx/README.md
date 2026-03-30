# stringsx

String and string-slice utility functions.

## Functions

| Function | Description |
|----------|-------------|
| `Diff(base, exclude)` | Elements in `base` not in `exclude` |
| `Include(base, include)` | Elements in both `base` and `include` |
| `Unique(ss)` | Deduplicated copy (preserves first occurrence order) |
| `Contains(list, s)` | Exact membership check |
| `ContainsEqualFold(list, s)` | Case-insensitive membership check |
| `Index(array, s)` | Index of `s` in `array`, or `-1` |
| `Filter(list, s)` | New slice with all occurrences of `s` removed |
| `Add(list, s)` | Append `s` only if not already present |
| `Reverse(s)` | UTF-8 aware string reversal |
| `DecodeBase64(s)` | Standard base64 decode |

## Examples

### Set operations

```go
all     := []string{"read", "write", "admin", "audit"}
revoked := []string{"admin", "audit"}

remaining := stringsx.Diff(all, revoked)
// ["read", "write"]

overlap := stringsx.Include(all, []string{"write", "delete"})
// ["write"]
```

### Deduplication and membership

```go
tags := []string{"go", "rust", "go", "python", "rust"}

unique := stringsx.Unique(tags)
// ["go", "rust", "python"]

stringsx.Contains(tags, "go")              // true
stringsx.ContainsEqualFold(tags, "PYTHON") // true
```

### Filtering and building slices

```go
roles := []string{"user", "guest", "user", "admin"}

filtered := stringsx.Filter(roles, "guest")
// ["user", "user", "admin"]

roles = stringsx.Add(roles, "superadmin")
// ["user", "guest", "user", "admin", "superadmin"]

roles = stringsx.Add(roles, "admin") // no-op, already present
```

### String utilities

```go
stringsx.Reverse("Hello, 世界") // "界世 ,olleH"

data, err := stringsx.DecodeBase64("SGVsbG8=") // []byte("Hello")
```
