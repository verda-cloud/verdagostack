package gin

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestObservability_InjectTraceIDOnly(t *testing.T) {
	router := gin.New()
	router.Use(Observability(WithTraceInjection(InjectTraceIDOnly)))
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Without an active span, the trace ID header should not be set
	// (zero SpanContext is not valid).
	if h := w.Header().Get(TraceIDHeaderKey); h != "" {
		t.Fatalf("expected no %s header without active span, got %q", TraceIDHeaderKey, h)
	}
}

func TestObservability_InjectW3C(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	cfg := trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	}
	sc := trace.NewSpanContext(cfg)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := trace.ContextWithRemoteSpanContext(c.Request.Context(), sc)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(Observability(WithTraceInjection(InjectW3CTraceContext)))
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	traceparent := w.Header().Get(TraceParentHeaderKey)
	if traceparent == "" {
		t.Fatal("expected traceparent header")
	}
	if !strings.HasPrefix(traceparent, "00-") {
		t.Fatalf("traceparent should start with 00-, got %q", traceparent)
	}
	if !strings.HasSuffix(traceparent, "-01") {
		t.Fatalf("traceparent should end with -01 (sampled), got %q", traceparent)
	}
}

func TestObservability_InjectBoth(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled, Remote: true,
	})

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := trace.ContextWithRemoteSpanContext(c.Request.Context(), sc)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(Observability(WithTraceInjection(InjectBoth)))
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Header().Get(TraceParentHeaderKey) == "" {
		t.Fatal("expected traceparent header for InjectBoth")
	}
	if w.Header().Get(TraceIDHeaderKey) == "" {
		t.Fatal("expected X-Trace-Id header for InjectBoth")
	}
}

func TestObservability_InjectNone(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled, Remote: true,
	})

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := trace.ContextWithRemoteSpanContext(c.Request.Context(), sc)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(Observability(WithTraceInjection(InjectNone)))
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Header().Get(TraceParentHeaderKey) != "" {
		t.Fatal("expected no traceparent for InjectNone")
	}
	if w.Header().Get(TraceIDHeaderKey) != "" {
		t.Fatal("expected no X-Trace-Id for InjectNone")
	}
}

func TestObservability_SkipPaths(t *testing.T) {
	router := gin.New()
	router.Use(Observability(WithSkipPaths("/healthz", "/ready")))
	router.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestObservability_CustomLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	router := gin.New()
	router.Use(Observability(WithLogger(logger)))
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if !strings.Contains(buf.String(), "HTTP request completed") {
		t.Fatalf("expected log output with custom logger, got %q", buf.String())
	}
}

func TestObservability_BodyCapture(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	router := gin.New()
	router.Use(Observability(WithLogger(logger)))
	router.POST("/echo", func(c *gin.Context) {
		body := make([]byte, 256)
		n, _ := c.Request.Body.Read(body)
		c.String(http.StatusOK, string(body[:n]))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"msg":"hello"}`))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestObservability_DisableBodyLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	router := gin.New()
	router.Use(Observability(WithLogger(logger), WithDisableBodyLog()))
	router.POST("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"msg":"hello"}`))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestObservability_CustomTraceHeader(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled, Remote: true,
	})

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := trace.ContextWithRemoteSpanContext(c.Request.Context(), sc)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(Observability(WithTraceInjection(InjectTraceIDOnly), WithCustomTraceHeader("X-My-Trace")))
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if h := w.Header().Get("X-My-Trace"); h == "" {
		t.Fatal("expected custom trace header X-My-Trace")
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		path, method string
		patterns     []string
		want         bool
	}{
		{"/healthz", "GET", []string{"/healthz"}, true},
		{"/api/v1/users", "GET", []string{"/healthz"}, false},
		{"/api/v1/users", "GET", []string{"/api/*"}, true},
		{"/metrics", "GET", []string{"GET /metrics"}, true},
		{"/metrics", "POST", []string{"GET /metrics"}, false},
		{"/api/v1/health/ready", "GET", []string{"/api/"}, true},
		{"/test.json", "GET", []string{"*.json"}, true},
	}
	for _, tt := range tests {
		got := shouldSkipPath(tt.path, tt.method, tt.patterns)
		if got != tt.want {
			t.Errorf("shouldSkipPath(%q, %q, %v) = %v, want %v", tt.path, tt.method, tt.patterns, got, tt.want)
		}
	}
}

func TestConvenienceConstructors(t *testing.T) {
	// Just verify they return non-nil handlers without panicking.
	if ObservabilityWithW3CTraceContext() == nil {
		t.Fatal("ObservabilityWithW3CTraceContext returned nil")
	}
	if ObservabilityWithTraceID() == nil {
		t.Fatal("ObservabilityWithTraceID returned nil")
	}
	if ObservabilityWithCustomHeader("X-Test") == nil {
		t.Fatal("ObservabilityWithCustomHeader returned nil")
	}
	if ObservabilitySkipMetrics() == nil {
		t.Fatal("ObservabilitySkipMetrics returned nil")
	}
	if ObservabilityWithSkipPaths("/foo") == nil {
		t.Fatal("ObservabilityWithSkipPaths returned nil")
	}
}

// Suppress unused import warning for context.
var _ = context.Background
