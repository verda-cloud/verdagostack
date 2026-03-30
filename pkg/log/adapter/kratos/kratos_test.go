package kratos

import (
	"testing"

	krtlog "github.com/go-kratos/kratos/v2/log"

	"github.com/verda-cloud/verdagostack/pkg/log/empty"
)

func TestKratosLogger_ImplementsInterface(t *testing.T) {
	l := New(empty.NewLogger())
	var _ krtlog.Logger = l
}

func TestKratosLogger_Log_AllLevels(t *testing.T) {
	levels := []krtlog.Level{
		krtlog.LevelDebug,
		krtlog.LevelInfo,
		krtlog.LevelWarn,
		krtlog.LevelError,
	}

	l := New(empty.NewLogger())
	for _, level := range levels {
		if err := l.Log(level, "key", "value"); err != nil {
			t.Errorf("Log(%v) returned error: %v", level, err)
		}
	}
}

func TestKratosLogger_Log_EmptyKeyvals(t *testing.T) {
	l := New(empty.NewLogger())
	// Should not panic, just warn
	if err := l.Log(krtlog.LevelInfo); err != nil {
		t.Errorf("Log with empty keyvals returned error: %v", err)
	}
}

func TestKratosLogger_Log_OddKeyvals(t *testing.T) {
	l := New(empty.NewLogger())
	// Odd number of keyvals — should warn but not panic or error
	if err := l.Log(krtlog.LevelInfo, "orphan_key"); err != nil {
		t.Errorf("Log with odd keyvals returned error: %v", err)
	}
}

func TestKratosLogger_Log_MultipleKVPairs(t *testing.T) {
	l := New(empty.NewLogger())
	err := l.Log(krtlog.LevelInfo, "method", "GET", "path", "/api", "status", 200)
	if err != nil {
		t.Errorf("Log with multiple KV pairs returned error: %v", err)
	}
}
