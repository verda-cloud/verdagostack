package empty

import (
	"context"
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/log"
)

func TestEmptyLogger_ImplementsInterface(t *testing.T) {
	var logger log.Logger = NewLogger()
	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}
}

func TestEmptyLogger_NoPanic(t *testing.T) {
	l := NewLogger()

	// None of these should panic
	l.Debugf("test %s", "debug")
	l.Infof("test %s", "info")
	l.Warnf("test %s", "warn")
	l.Errorf("test %s", "error")

	l.Debugw("test", "key", "val")
	l.Infow("test", "key", "val")
	l.Warnw("test", "key", "val")
	l.Errorw("test", "key", "val")

	l.Sync()
}

func TestEmptyLogger_With_ReturnsSelf(t *testing.T) {
	l := NewLogger()
	child := l.With("key", "val")
	if child != l {
		t.Error("With() should return the same EmptyLogger instance")
	}
}

func TestEmptyLogger_W_ReturnsSelf(t *testing.T) {
	l := NewLogger()
	ctxLogger := l.W(context.Background())
	if ctxLogger != l {
		t.Error("W() should return the same EmptyLogger instance")
	}
}

func TestEmptyLogger_AddCallerSkip_ReturnsSelf(t *testing.T) {
	l := NewLogger()
	skipped := l.AddCallerSkip(5)
	if skipped != l {
		t.Error("AddCallerSkip() should return the same EmptyLogger instance")
	}
}

func TestEmptyLogger_Chaining(t *testing.T) {
	l := NewLogger()
	// Chaining should work without panics
	l.With("a", "b").W(context.Background()).AddCallerSkip(1).Infow("chained")
}
