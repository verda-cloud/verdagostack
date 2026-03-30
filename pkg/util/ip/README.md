# ip

Utilities for obtaining local and remote IP addresses.

## Functions

| Function | Description |
|----------|-------------|
| `GetLocalIP()` | First non-loopback IPv4 address of the host (falls back to `127.0.0.1`) |
| `RemoteIP(req)` | Client IP from HTTP request, checking proxy headers before `RemoteAddr` |

## Proxy header priority

`RemoteIP` checks headers in this order:

1. `X-Client-IP`
2. `X-Real-IP`
3. `X-Forwarded-For`
4. `req.RemoteAddr` (fallback)

## Examples

### Get local machine IP

```go
localIP := ip.GetLocalIP()
// e.g. "192.168.1.42"
```

### Extract client IP in an HTTP handler

```go
func handler(w http.ResponseWriter, r *http.Request) {
    clientIP := ip.RemoteIP(r)
    slog.Info("request received", "client_ip", clientIP)
}
```
