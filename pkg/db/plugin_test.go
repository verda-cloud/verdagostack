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
	if name := p.Name(); name != "verdagostack:trace" {
		t.Errorf("expected name verdagostack:trace, got %s", name)
	}
}
