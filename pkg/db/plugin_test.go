package db

import (
	"testing"

	"gorm.io/gorm"
)

func TestTracePlugin_ImplementsPlugin(t *testing.T) {
	var _ gorm.Plugin = TracePlugin{}
}

func TestTracePlugin_Name(t *testing.T) {
	p := TracePlugin{}
	if name := p.Name(); name != "verdastack:trace" {
		t.Errorf("expected name verdastack:trace, got %s", name)
	}
}
