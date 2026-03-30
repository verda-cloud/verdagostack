package log

import (
	"context"

	"go.uber.org/zap"
)

// ContextExtractors maps field names to functions that extract their values
// from a context.Context. Registered extractors are invoked by W() to enrich
// log entries with request-scoped metadata (e.g., request ID, trace ID).
type ContextExtractors map[string]func(context.Context) string

// WithContextExtractor returns an Option that registers the given extractors.
// Multiple calls are additive; later registrations for the same key overwrite
// earlier ones.
func WithContextExtractor(extractors ContextExtractors) Option {
	return func(l *zapLogger) {
		for k, v := range extractors {
			l.contextExtractors[k] = v
		}
	}
}

// W returns a new Logger whose output is enriched with fields extracted
// from ctx by the registered ContextExtractors.
func (l *zapLogger) W(ctx context.Context) Logger {
	lc := l.clone()
	for fieldName, extractor := range l.contextExtractors {
		if val := extractor(ctx); val != "" {
			lc.z = lc.z.With(zap.String(fieldName, val))
		}
	}
	return lc
}
