// Package gin provides a Gin framework adapter that routes Gin's request
// logging and internal debug output through a verdastack log.Logger.
//
// Two integration points are provided:
//
//   - NewWriter: an io.Writer bridge for gin.DefaultWriter / gin.DefaultErrorWriter,
//     capturing Gin's internal debug prints (route registration, etc.).
//
//   - Middleware: a drop-in replacement for gin.Logger() that produces structured
//     request logs with method, path, status, latency, client IP, and any
//     context-extracted fields (request ID, trace ID, etc.) via Logger.W(ctx).
package gin

import (
	"io"
	"strings"
	"time"

	ginfw "github.com/gin-gonic/gin"

	"github.com/verda-cloud/verdagostack/pkg/log"
)

// --- io.Writer bridge ---

// Writer returns an io.Writer that logs each Write call as a single message
// at the specified level. Use it to redirect Gin's built-in output:
//
//	gin.DefaultWriter = logadaptergin.Writer(log.Default(), logadaptergin.LevelInfo)
//	gin.DefaultErrorWriter = logadaptergin.Writer(log.Default(), logadaptergin.LevelError)
func Writer(logger log.Logger, level Level) io.Writer {
	return &logWriter{logger: logger.AddCallerSkip(1), level: level}
}

// Level controls which log method the Writer bridge delegates to.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type logWriter struct {
	logger log.Logger
	level  Level
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	if msg == "" {
		return len(p), nil
	}

	switch w.level {
	case LevelDebug:
		w.logger.Debugf("%s", msg)
	case LevelInfo:
		w.logger.Infof("%s", msg)
	case LevelWarn:
		w.logger.Warnf("%s", msg)
	case LevelError:
		w.logger.Errorf("%s", msg)
	default:
		w.logger.Infof("%s", msg)
	}
	return len(p), nil
}

// --- Structured request logging middleware ---

// MiddlewareConfig holds optional configuration for the logging middleware.
type MiddlewareConfig struct {
	// SkipPaths is a set of URL paths that should not be logged.
	// Useful for health-check endpoints that generate noise.
	SkipPaths []string

	// SlowThreshold marks requests slower than this as "slow" at warn level.
	// Zero means no slow-request detection.
	SlowThreshold time.Duration
}

// Middleware returns a gin.HandlerFunc that logs every HTTP request as a
// structured log entry through the provided Logger.
//
// It replaces gin.Logger() and produces fields: method, path, status,
// latency_ms, client_ip, user_agent, and error (if any).
// Context-extracted fields (request ID, trace ID) are included automatically
// via Logger.W(c.Request.Context()).
//
//	router := gin.New()
//	router.Use(logadaptergin.Middleware(log.Default(), nil))
func Middleware(logger log.Logger, cfg *MiddlewareConfig) ginfw.HandlerFunc {
	skipSet := make(map[string]struct{})
	var slowThreshold time.Duration
	if cfg != nil {
		for _, p := range cfg.SkipPaths {
			skipSet[p] = struct{}{}
		}
		slowThreshold = cfg.SlowThreshold
	}

	return func(c *ginfw.Context) {
		path := c.Request.URL.Path
		if _, skip := skipSet[path]; skip {
			c.Next()
			return
		}

		start := time.Now()

		// Process the request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		rawQuery := c.Request.URL.RawQuery

		fullPath := path
		if rawQuery != "" {
			fullPath = path + "?" + rawQuery
		}

		fields := []any{
			"method", method,
			"path", fullPath,
			"status", status,
			"latency_ms", float64(latency.Nanoseconds()) / 1e6,
			"client_ip", clientIP,
			"user_agent", userAgent,
		}

		if len(c.Errors) > 0 {
			fields = append(fields, "errors", c.Errors.String())
		}

		// Use W(ctx) to pick up context-extracted fields (request ID, trace ID, etc.)
		l := logger.W(c.Request.Context())

		switch {
		case status >= 500:
			l.Errorw("HTTP request", fields...)
		case status >= 400:
			l.Warnw("HTTP request", fields...)
		case slowThreshold > 0 && latency > slowThreshold:
			fields = append(fields, "slow", true, "threshold_ms", float64(slowThreshold.Nanoseconds())/1e6)
			l.Warnw("HTTP request", fields...)
		default:
			l.Infow("HTTP request", fields...)
		}
	}
}
