# retry

Exponential backoff, polling, and periodic execution helpers using only the Go standard library.

## Functions

| Function | Description |
|----------|-------------|
| `Retry(fn, cfg)` | Exponential backoff — retries until success, error, or attempts exhausted |
| `Poll(ctx, interval, timeout, fn)` | Fixed-interval polling (first check after one interval) |
| `PollImmediate(ctx, interval, timeout, fn)` | Like `Poll` but checks immediately first |
| `RunImmediatelyThenPeriod(ctx, fn, period)` | Run once, then repeat in background goroutine |
| `DefaultBackoff()` | Sensible default `BackoffConfig` (10 steps, 5s, 1.25×, jitter 1.0) |

## ConditionFunc return values

| Return | Meaning |
|--------|---------|
| `(false, nil)` | Keep retrying |
| `(true, nil)` | Success — stop |
| `(false, err)` | Permanent failure — stop immediately |

## Examples

### HTTP call with retry

```go
var resp *http.Response
err := retry.Retry(func() (bool, error) {
    var reqErr error
    resp, reqErr = http.Get("https://api.example.com/data")
    if reqErr != nil {
        return false, nil // transient network error, keep retrying
    }
    if resp.StatusCode >= 500 {
        resp.Body.Close()
        return false, nil // server error, retry
    }
    if resp.StatusCode == http.StatusUnauthorized {
        resp.Body.Close()
        return false, fmt.Errorf("401 unauthorized") // permanent, stop
    }
    return true, nil // success — caller reads resp.Body
}, retry.BackoffConfig{Steps: 5, Duration: time.Second, Factor: 2, Jitter: 0.5})
```

### Database connection with retry

```go
var db *sql.DB
err := retry.Retry(func() (bool, error) {
    var connErr error
    db, connErr = sql.Open("postgres", dsn)
    if connErr != nil {
        return false, nil
    }
    if err := db.Ping(); err != nil {
        db.Close()
        return false, nil
    }
    return true, nil
}, retry.DefaultBackoff())
```

### Wait for a dependency to become healthy

```go
err := retry.PollImmediate(ctx, 2*time.Second, 30*time.Second, func() (bool, error) {
    resp, err := http.Get("http://localhost:8080/healthz")
    if err != nil {
        return false, nil
    }
    resp.Body.Close()
    return resp.StatusCode == 200, nil
})
```

### Periodic background refresh

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := retry.RunImmediatelyThenPeriod(ctx, func(ctx context.Context) error {
    return refreshCache(ctx)
}, 5*time.Minute)
```
