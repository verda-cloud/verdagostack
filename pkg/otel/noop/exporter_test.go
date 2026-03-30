package noop

import (
	"context"
	"testing"
)

func TestExporter_ExportSpans(t *testing.T) {
	e := NewExporter()
	if err := e.ExportSpans(context.Background(), nil); err != nil {
		t.Errorf("ExportSpans returned unexpected error: %v", err)
	}
}

func TestExporter_Shutdown(t *testing.T) {
	e := NewExporter()
	if err := e.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
}
