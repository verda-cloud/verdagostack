// Package gin provides observability middleware for the Gin HTTP framework.
// It extracts OTel trace context from requests and injects trace headers into
// responses, with configurable injection modes and structured request logging.
package gin

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

const (
	TraceParentHeaderKey = "traceparent"
	TraceIDHeaderKey     = "X-Trace-Id"
	RequestIDHeaderKey   = "X-Request-Id"
	TraceStateHeaderKey  = "tracestate"
)

// TraceInjectionMode defines how trace information is injected into responses.
type TraceInjectionMode int

const (
	InjectW3CTraceContext TraceInjectionMode = iota
	InjectTraceIDOnly
	InjectBoth
	InjectNone
)

// ObservabilityOptions configures trace injection and request logging.
type ObservabilityOptions struct {
	TraceInjectionMode TraceInjectionMode
	CustomTraceHeader  string
	SkipPaths          []string
	Logger             *slog.Logger
	DisableBodyLog     bool
}

// Option is a functional option for configuring the middleware.
type Option func(*ObservabilityOptions)

// WithTraceInjection configures the trace injection mode.
func WithTraceInjection(mode TraceInjectionMode) Option {
	return func(o *ObservabilityOptions) { o.TraceInjectionMode = mode }
}

// WithCustomTraceHeader sets a custom header name for the trace ID.
func WithCustomTraceHeader(headerName string) Option {
	return func(o *ObservabilityOptions) { o.CustomTraceHeader = headerName }
}

// WithSkipPaths adds paths that should skip logging. Supports exact match,
// prefix match (trailing /), wildcards (*), and method-specific patterns
// (e.g., "GET /metrics").
func WithSkipPaths(paths ...string) Option {
	return func(o *ObservabilityOptions) { o.SkipPaths = append(o.SkipPaths, paths...) }
}

// WithLogger sets a custom slog.Logger. Defaults to slog.Default().
func WithLogger(logger *slog.Logger) Option {
	return func(o *ObservabilityOptions) {
		if logger != nil {
			o.Logger = logger
		}
	}
}

// WithDisableBodyLog prevents logging of request/response bodies even at debug level.
func WithDisableBodyLog() Option {
	return func(o *ObservabilityOptions) { o.DisableBodyLog = true }
}

// WithSkipMetrics skips common infrastructure endpoints like /healthz, /metrics, /readyz.
func WithSkipMetrics() Option {
	return func(o *ObservabilityOptions) {
		o.SkipPaths = append(o.SkipPaths,
			"/health", "/healthz", "/health/*",
			"/ready", "/readyz", "/readiness",
			"/live", "/liveness",
			"/metrics", "/prometheus",
			"/status", "/ping", "/version", "/info",
			"/favicon.ico", "/robots.txt",
		)
	}
}

// Observability returns a Gin middleware that injects trace headers into
// responses and logs structured request data with trace correlation.
// Assumes an OTel tracing middleware (e.g., otelgin) has already created spans.
func Observability(opts ...Option) gin.HandlerFunc {
	config := &ObservabilityOptions{
		TraceInjectionMode: InjectTraceIDOnly,
		SkipPaths:          []string{"/metrics"},
		Logger:             slog.Default(),
	}
	for _, opt := range opts {
		opt(config)
	}

	return func(c *gin.Context) {
		if shouldSkipPath(c.Request.URL.Path, c.Request.Method, config.SkipPaths) {
			c.Next()
			return
		}

		start := time.Now()
		ctx := c.Request.Context()
		spanCtx := trace.SpanContextFromContext(ctx)

		injectTraceHeaders(c, spanCtx, config)

		isDebugLevel := config.Logger.Enabled(ctx, slog.LevelDebug)
		shouldLogBody := isDebugLevel && !config.DisableBodyLog

		var requestBody string
		var responseBuffer bytes.Buffer

		if shouldLogBody && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		if shouldLogBody {
			c.Writer = &bodyCaptureWriter{ResponseWriter: c.Writer, body: &responseBuffer}
		}

		c.Next()

		duration := time.Since(start).Seconds()

		httpData := map[string]any{
			"request": map[string]any{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"id":     spanCtx.TraceID().String(),
			},
			"response": map[string]any{
				"status_code": c.Writer.Status(),
			},
		}

		if shouldLogBody {
			httpData["request"].(map[string]any)["body"] = map[string]any{
				"content": requestBody,
				"bytes":   len(requestBody),
			}
			httpData["response"].(map[string]any)["body"] = map[string]any{
				"content": responseBuffer.String(),
				"bytes":   responseBuffer.Len(),
			}
		}

		logLevel := slog.LevelInfo
		if isDebugLevel {
			logLevel = slog.LevelDebug
		}

		config.Logger.Log(ctx, logLevel, "HTTP request completed",
			"event", map[string]any{"duration": duration},
			"source", map[string]any{"ip": c.ClientIP()},
			"http", httpData,
			"user_agent", map[string]any{"original": c.Request.UserAgent()},
		)
	}
}

func shouldSkipPath(path, method string, skipPaths []string) bool {
	for _, pattern := range skipPaths {
		if matchPath(path, method, pattern) {
			return true
		}
	}
	return false
}

func matchPath(requestPath, method, pattern string) bool {
	if strings.Contains(pattern, " ") {
		parts := strings.SplitN(pattern, " ", 2)
		if len(parts) == 2 {
			if strings.ToUpper(strings.TrimSpace(parts[0])) != strings.ToUpper(method) {
				return false
			}
			return matchPathPattern(requestPath, strings.TrimSpace(parts[1]))
		}
	}
	return matchPathPattern(requestPath, pattern)
}

func matchPathPattern(path, pattern string) bool {
	if path == pattern {
		return true
	}
	if strings.Contains(pattern, "*") {
		return matchWildcard(path, pattern)
	}
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern)
	}
	return false
}

func matchWildcard(text, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(text, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(text, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(text, pattern[:len(pattern)-1])
	}
	return text == pattern
}

func injectTraceHeaders(c *gin.Context, spanCtx trace.SpanContext, config *ObservabilityOptions) {
	if !spanCtx.IsValid() {
		return
	}

	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()
	traceFlags := "01"
	if !spanCtx.IsSampled() {
		traceFlags = "00"
	}

	switch config.TraceInjectionMode {
	case InjectW3CTraceContext:
		c.Header(TraceParentHeaderKey, fmt.Sprintf("00-%s-%s-%s", traceID, spanID, traceFlags))
	case InjectTraceIDOnly:
		header := TraceIDHeaderKey
		if config.CustomTraceHeader != "" {
			header = config.CustomTraceHeader
		}
		c.Header(header, traceID)
	case InjectBoth:
		c.Header(TraceParentHeaderKey, fmt.Sprintf("00-%s-%s-%s", traceID, spanID, traceFlags))
		header := TraceIDHeaderKey
		if config.CustomTraceHeader != "" {
			header = config.CustomTraceHeader
		}
		c.Header(header, traceID)
	case InjectNone:
	}
}

// Convenience constructors for common configurations.

func ObservabilityWithW3CTraceContext() gin.HandlerFunc {
	return Observability(WithTraceInjection(InjectW3CTraceContext))
}

func ObservabilityWithTraceID() gin.HandlerFunc {
	return Observability(WithTraceInjection(InjectTraceIDOnly))
}

func ObservabilityWithCustomHeader(headerName string) gin.HandlerFunc {
	return Observability(WithTraceInjection(InjectTraceIDOnly), WithCustomTraceHeader(headerName))
}

func ObservabilitySkipMetrics() gin.HandlerFunc {
	return Observability(WithSkipMetrics())
}

func ObservabilityWithSkipPaths(paths ...string) gin.HandlerFunc {
	return Observability(WithSkipPaths(paths...))
}

// bodyCaptureWriter duplicates response body writes to a buffer for logging.
type bodyCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
