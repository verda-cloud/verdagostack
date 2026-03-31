package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ginfw "github.com/gin-gonic/gin"

	"github.com/verda-cloud/verdagostack/pkg/log"
)

// --- recording logger for test assertions ---

type entry struct {
	level  string
	msg    string
	kvs    []any
	format string
	args   []any
}

type recorder struct {
	entries []entry
}

func (r *recorder) reset() { r.entries = nil }
func (r *recorder) Debugf(f string, a ...any) {
	r.entries = append(r.entries, entry{level: "debug", format: f, args: a})
}
func (r *recorder) Infof(f string, a ...any) {
	r.entries = append(r.entries, entry{level: "info", format: f, args: a})
}
func (r *recorder) Warnf(f string, a ...any) {
	r.entries = append(r.entries, entry{level: "warn", format: f, args: a})
}
func (r *recorder) Errorf(f string, a ...any) {
	r.entries = append(r.entries, entry{level: "error", format: f, args: a})
}
func (r *recorder) Panicf(_ string, _ ...any) {}
func (r *recorder) Fatalf(_ string, _ ...any) {}
func (r *recorder) Debugw(m string, kv ...any) {
	r.entries = append(r.entries, entry{level: "debug", msg: m, kvs: kv})
}
func (r *recorder) Infow(m string, kv ...any) {
	r.entries = append(r.entries, entry{level: "info", msg: m, kvs: kv})
}
func (r *recorder) Warnw(m string, kv ...any) {
	r.entries = append(r.entries, entry{level: "warn", msg: m, kvs: kv})
}
func (r *recorder) Errorw(m string, kv ...any) {
	r.entries = append(r.entries, entry{level: "error", msg: m, kvs: kv})
}
func (r *recorder) Panicw(_ string, _ ...any)      {}
func (r *recorder) Fatalw(_ string, _ ...any)      {}
func (r *recorder) W(_ context.Context) log.Logger { return r }
func (r *recorder) With(_ ...any) log.Logger       { return r }
func (r *recorder) AddCallerSkip(_ int) log.Logger { return r }
func (r *recorder) Sync()                          {}

var _ log.Logger = (*recorder)(nil)

func (r *recorder) lastEntry() entry {
	if len(r.entries) == 0 {
		return entry{}
	}
	return r.entries[len(r.entries)-1]
}

func (r *recorder) findKV(key string) (any, bool) {
	e := r.lastEntry()
	for i := 0; i+1 < len(e.kvs); i += 2 {
		if e.kvs[i] == key {
			return e.kvs[i+1], true
		}
	}
	return nil, false
}

// --- Writer tests ---

func TestWriter_Write_InfoLevel(t *testing.T) {
	rec := &recorder{}
	w := Writer(rec, LevelInfo)

	n, err := w.Write([]byte("gin debug output\n"))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len("gin debug output\n") {
		t.Errorf("expected n=%d, got %d", len("gin debug output\n"), n)
	}
	if len(rec.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rec.entries))
	}
	if rec.entries[0].level != "info" {
		t.Errorf("expected level 'info', got %q", rec.entries[0].level)
	}
}

func TestWriter_Write_ErrorLevel(t *testing.T) {
	rec := &recorder{}
	w := Writer(rec, LevelError)

	w.Write([]byte("something broke"))
	if rec.lastEntry().level != "error" {
		t.Errorf("expected level 'error', got %q", rec.lastEntry().level)
	}
}

func TestWriter_Write_DebugLevel(t *testing.T) {
	rec := &recorder{}
	w := Writer(rec, LevelDebug)

	w.Write([]byte("debug line"))
	if rec.lastEntry().level != "debug" {
		t.Errorf("expected level 'debug', got %q", rec.lastEntry().level)
	}
}

func TestWriter_Write_WarnLevel(t *testing.T) {
	rec := &recorder{}
	w := Writer(rec, LevelWarn)

	w.Write([]byte("warn line"))
	if rec.lastEntry().level != "warn" {
		t.Errorf("expected level 'warn', got %q", rec.lastEntry().level)
	}
}

func TestWriter_Write_TrimsTrailingNewline(t *testing.T) {
	rec := &recorder{}
	w := Writer(rec, LevelInfo)

	w.Write([]byte("no trailing newline\n\n"))
	if len(rec.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rec.entries))
	}
}

func TestWriter_Write_EmptySkipped(t *testing.T) {
	rec := &recorder{}
	w := Writer(rec, LevelInfo)

	w.Write([]byte("\n"))
	if len(rec.entries) != 0 {
		t.Error("empty (newline-only) writes should be skipped")
	}
}

func TestWriter_AsGinDefaultWriter(t *testing.T) {
	rec := &recorder{}
	ginfw.DefaultWriter = Writer(rec, LevelInfo)
	defer func() { ginfw.DefaultWriter = nil }()

	// Writing through gin's default writer should reach our logger
	ginfw.DefaultWriter.Write([]byte("test via gin.DefaultWriter"))
	if len(rec.entries) == 0 {
		t.Error("expected log entry from gin.DefaultWriter")
	}
}

// --- Middleware tests ---

func init() {
	ginfw.SetMode(ginfw.TestMode)
}

func setupRouter(logger log.Logger, cfg *MiddlewareConfig) *ginfw.Engine {
	r := ginfw.New()
	r.Use(Middleware(logger, cfg))
	r.GET("/ok", func(c *ginfw.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/not-found", func(c *ginfw.Context) { c.String(http.StatusNotFound, "nope") })
	r.GET("/error", func(c *ginfw.Context) { c.String(http.StatusInternalServerError, "fail") })
	r.GET("/health", func(c *ginfw.Context) { c.String(http.StatusOK, "healthy") })
	r.GET("/slow", func(c *ginfw.Context) {
		time.Sleep(50 * time.Millisecond)
		c.String(http.StatusOK, "slow")
	})
	return r
}

func TestMiddleware_200_LogsInfoLevel(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	if len(rec.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(rec.entries))
	}
	e := rec.lastEntry()
	if e.level != "info" {
		t.Errorf("expected 'info' for 200, got %q", e.level)
	}
	if e.msg != "HTTP request" {
		t.Errorf("expected msg 'HTTP request', got %q", e.msg)
	}
	if v, ok := rec.findKV("method"); !ok || v != "GET" {
		t.Errorf("expected method=GET, got %v", v)
	}
	if v, ok := rec.findKV("path"); !ok || v != "/ok" {
		t.Errorf("expected path=/ok, got %v", v)
	}
	if v, ok := rec.findKV("status"); !ok || v != 200 {
		t.Errorf("expected status=200, got %v", v)
	}
}

func TestMiddleware_404_LogsWarnLevel(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/not-found", nil))

	if rec.lastEntry().level != "warn" {
		t.Errorf("expected 'warn' for 404, got %q", rec.lastEntry().level)
	}
}

func TestMiddleware_500_LogsErrorLevel(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/error", nil))

	if rec.lastEntry().level != "error" {
		t.Errorf("expected 'error' for 500, got %q", rec.lastEntry().level)
	}
}

func TestMiddleware_IncludesLatency(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/ok", nil))

	v, ok := rec.findKV("latency_ms")
	if !ok {
		t.Fatal("expected latency_ms field")
	}
	lat, ok := v.(float64)
	if !ok || lat < 0 {
		t.Errorf("expected non-negative latency, got %v", v)
	}
}

func TestMiddleware_IncludesUserAgent(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ok", nil)
	req.Header.Set("User-Agent", "verdagostack-test/1.0")
	r.ServeHTTP(w, req)

	v, ok := rec.findKV("user_agent")
	if !ok || v != "verdagostack-test/1.0" {
		t.Errorf("expected user_agent 'verdagostack-test/1.0', got %v", v)
	}
}

func TestMiddleware_IncludesQueryString(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/ok?page=2&limit=10", nil))

	v, ok := rec.findKV("path")
	if !ok {
		t.Fatal("expected path field")
	}
	path, _ := v.(string)
	if !strings.Contains(path, "page=2") || !strings.Contains(path, "limit=10") {
		t.Errorf("expected query string in path, got %q", path)
	}
}

func TestMiddleware_SkipPaths(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, &MiddlewareConfig{
		SkipPaths: []string{"/health"},
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))

	if len(rec.entries) != 0 {
		t.Error("expected /health to be skipped, but got log entries")
	}
}

func TestMiddleware_SkipPaths_DoesNotSkipOthers(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, &MiddlewareConfig{
		SkipPaths: []string{"/health"},
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/ok", nil))

	if len(rec.entries) != 1 {
		t.Errorf("expected 1 entry for /ok, got %d", len(rec.entries))
	}
}

func TestMiddleware_SlowThreshold(t *testing.T) {
	rec := &recorder{}
	r := setupRouter(rec, &MiddlewareConfig{
		SlowThreshold: 10 * time.Millisecond,
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/slow", nil))

	e := rec.lastEntry()
	if e.level != "warn" {
		t.Errorf("expected 'warn' for slow request, got %q", e.level)
	}
	if v, ok := rec.findKV("slow"); !ok || v != true {
		t.Error("expected slow=true for slow request")
	}
}

func TestMiddleware_GinErrors(t *testing.T) {
	rec := &recorder{}
	r := ginfw.New()
	r.Use(Middleware(rec, nil))
	r.GET("/with-error", func(c *ginfw.Context) {
		_ = c.Error(http.ErrBodyNotAllowed)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/with-error", nil))

	v, ok := rec.findKV("errors")
	if !ok {
		t.Fatal("expected errors field when c.Error is set")
	}
	errStr, _ := v.(string)
	if errStr == "" {
		t.Error("expected non-empty errors string")
	}
}

func TestMiddleware_NilConfig(t *testing.T) {
	rec := &recorder{}
	r := ginfw.New()
	r.Use(Middleware(rec, nil))
	r.GET("/ping", func(c *ginfw.Context) { c.String(http.StatusOK, "pong") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/ping", nil))

	if len(rec.entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(rec.entries))
	}
}

// --- Integration test: real Logger with JSON output ---

func TestMiddleware_Integration_JSONOutput(t *testing.T) {
	opts := &log.Options{
		Level:         "debug",
		Format:        "json",
		OutputPaths:   []string{"stdout"},
		DisableCaller: true,
	}
	logger := log.NewLogger(opts)

	r := ginfw.New()
	r.Use(Middleware(logger, nil))
	r.GET("/api/v1/users", func(c *ginfw.Context) {
		c.JSON(http.StatusOK, ginfw.H{"users": []string{}})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/users", nil))

	// If we got here without panic, the integration works.
	// Verify the response is correct.
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

// --- Writer integration: verify bytes.Buffer interop ---

func TestWriter_Integration_BytesBuffer(t *testing.T) {
	// Verify the Writer works as a generic io.Writer (e.g., for io.MultiWriter)
	rec := &recorder{}
	w := Writer(rec, LevelWarn)
	var buf bytes.Buffer

	// Write to both
	multi := multiWriter(&buf, w)
	multi.Write([]byte("dual output"))

	if buf.String() != "dual output" {
		t.Errorf("expected buffer to contain 'dual output', got %q", buf.String())
	}
	if len(rec.entries) != 1 || rec.entries[0].level != "warn" {
		t.Error("expected one warn entry in recorder")
	}
}

type dualWriter struct {
	a, b interface{ Write([]byte) (int, error) }
}

func multiWriter(a, b interface{ Write([]byte) (int, error) }) *dualWriter {
	return &dualWriter{a: a, b: b}
}

func (d *dualWriter) Write(p []byte) (int, error) {
	d.a.Write(p)
	return d.b.Write(p)
}
