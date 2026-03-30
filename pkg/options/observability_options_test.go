package options

import (
	"testing"
)

func TestNewObservabilityOptions_Defaults(t *testing.T) {
	opts := NewObservabilityOptions()
	if opts.BindAddress != "0.0.0.0:9090" {
		t.Errorf("expected bind address '0.0.0.0:9090', got %q", opts.BindAddress)
	}
	if opts.EnablePprof {
		t.Error("EnablePprof should default to false")
	}
}

func TestObservabilityOptions_Validate_Valid(t *testing.T) {
	opts := NewObservabilityOptions()
	errs := opts.Validate()
	if len(errs) > 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestObservabilityOptions_Validate_Nil(t *testing.T) {
	var opts *ObservabilityOptions
	errs := opts.Validate()
	if len(errs) > 0 {
		t.Errorf("nil options should validate without errors, got %v", errs)
	}
}

func TestObservabilityOptions_Validate_BadAddress(t *testing.T) {
	opts := &ObservabilityOptions{BindAddress: "bad-addr"}
	errs := opts.Validate()
	if len(errs) != 1 {
		t.Errorf("expected 1 validation error for bad address, got %d: %v", len(errs), errs)
	}
}

func TestObservabilityOptions_DefaultHandlers(t *testing.T) {
	opts := NewObservabilityOptions()

	if h := opts.DefaultHealthzHandler(); h == nil {
		t.Error("DefaultHealthzHandler should not return nil")
	}
	if h := opts.DefaultMetricsHandler(); h == nil {
		t.Error("DefaultMetricsHandler should not return nil")
	}
}
