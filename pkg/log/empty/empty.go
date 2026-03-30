// Package empty provides a no-op Logger implementation for use in tests
// and anywhere logging output should be silently discarded.
package empty

import (
	"context"

	"github.com/verda-cloud/verdagostack/pkg/log"
)

// EmptyLogger is a Logger that silently discards all output.
type EmptyLogger struct{}

var _ log.Logger = (*EmptyLogger)(nil)

// NewLogger returns a no-op Logger.
func NewLogger() log.Logger {
	return &EmptyLogger{}
}

func (l *EmptyLogger) Debugf(_ string, _ ...any) {}
func (l *EmptyLogger) Infof(_ string, _ ...any)  {}
func (l *EmptyLogger) Warnf(_ string, _ ...any)  {}
func (l *EmptyLogger) Errorf(_ string, _ ...any) {}
func (l *EmptyLogger) Panicf(_ string, _ ...any) {}
func (l *EmptyLogger) Fatalf(_ string, _ ...any) {}

func (l *EmptyLogger) Debugw(_ string, _ ...any) {}
func (l *EmptyLogger) Infow(_ string, _ ...any)  {}
func (l *EmptyLogger) Warnw(_ string, _ ...any)  {}
func (l *EmptyLogger) Errorw(_ string, _ ...any) {}
func (l *EmptyLogger) Panicw(_ string, _ ...any) {}
func (l *EmptyLogger) Fatalw(_ string, _ ...any) {}

func (l *EmptyLogger) W(_ context.Context) log.Logger  { return l }
func (l *EmptyLogger) With(_ ...any) log.Logger         { return l }
func (l *EmptyLogger) AddCallerSkip(_ int) log.Logger   { return l }
func (l *EmptyLogger) Sync()                            {}
