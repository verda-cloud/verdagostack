package errorsx

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(http.StatusBadRequest, "BadInput", "field %s is required", "name")
	if err.Code != http.StatusBadRequest {
		t.Errorf("expected code %d, got %d", http.StatusBadRequest, err.Code)
	}
	if err.Reason != "BadInput" {
		t.Errorf("expected reason 'BadInput', got %q", err.Reason)
	}
	if err.Message != "field name is required" {
		t.Errorf("expected formatted message, got %q", err.Message)
	}
}

func TestError_String(t *testing.T) {
	err := New(500, "Internal", "boom")
	s := err.Error()
	if s == "" {
		t.Fatal("Error() returned empty string")
	}
	for _, substr := range []string{"500", "Internal", "boom"} {
		if !contains(s, substr) {
			t.Errorf("Error() %q missing substring %q", s, substr)
		}
	}
}

func TestWithMessage_ReturnsNewCopy(t *testing.T) {
	original := ErrInternal
	modified := original.WithMessage("custom: %s", "oops")

	if modified == original {
		t.Fatal("WithMessage must return a new *ErrorX, not the same pointer")
	}
	if modified.Message != "custom: oops" {
		t.Errorf("expected message 'custom: oops', got %q", modified.Message)
	}
	if original.Message != "Internal server error." {
		t.Errorf("original was mutated: got %q", original.Message)
	}
}

func TestWithMetadata_ReturnsNewCopy(t *testing.T) {
	original := ErrNotFound
	modified := original.WithMetadata(map[string]string{"id": "123"})

	if modified == original {
		t.Fatal("WithMetadata must return a new *ErrorX")
	}
	if modified.Metadata["id"] != "123" {
		t.Errorf("expected metadata id '123', got %q", modified.Metadata["id"])
	}
	if original.Metadata != nil {
		t.Error("original metadata was mutated")
	}
}

func TestKV_ReturnsNewCopy(t *testing.T) {
	base := New(400, "Test", "test").WithMetadata(map[string]string{"a": "1"})
	extended := base.KV("b", "2")

	if extended == base {
		t.Fatal("KV must return a new *ErrorX")
	}
	if extended.Metadata["a"] != "1" {
		t.Error("KV should preserve existing metadata")
	}
	if extended.Metadata["b"] != "2" {
		t.Error("KV should add new key-value pair")
	}
	if _, exists := base.Metadata["b"]; exists {
		t.Error("KV mutated the original error's metadata")
	}
}

func TestKV_OddNumberOfArgs(t *testing.T) {
	err := New(400, "Test", "test")
	extended := err.KV("only-key")
	if len(extended.Metadata) != 0 {
		t.Error("KV with odd args should not produce partial pairs")
	}
}

func TestWithRequestID(t *testing.T) {
	err := ErrInternal.WithRequestID("req-abc")
	if err.Metadata["X-Request-ID"] != "req-abc" {
		t.Errorf("expected X-Request-ID 'req-abc', got %q", err.Metadata["X-Request-ID"])
	}
}

func TestIs_MatchesByCodeAndReason(t *testing.T) {
	sentinel := ErrNotFound
	instance := New(http.StatusNotFound, "NotFound", "user 42 not found")

	if !errors.Is(instance, sentinel) {
		t.Error("errors.Is should match by Code+Reason")
	}
}

func TestIs_DifferentReasonDoesNotMatch(t *testing.T) {
	a := New(http.StatusBadRequest, "ReasonA", "msg")
	b := New(http.StatusBadRequest, "ReasonB", "msg")

	if errors.Is(a, b) {
		t.Error("errors.Is should not match when Reason differs")
	}
}

func TestIs_NonErrorXTarget(t *testing.T) {
	err := ErrInternal
	if errors.Is(err, errors.New("unrelated")) {
		t.Error("Is should return false for non-ErrorX target")
	}
}

func TestCode_NilError(t *testing.T) {
	if c := Code(nil); c != http.StatusOK {
		t.Errorf("Code(nil) should be 200, got %d", c)
	}
}

func TestCode_ErrorXError(t *testing.T) {
	err := ErrPermissionDenied
	if c := Code(err); c != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, c)
	}
}

func TestReason_NilError(t *testing.T) {
	if r := Reason(nil); r != ErrInternal.Reason {
		t.Errorf("Reason(nil) should be %q, got %q", ErrInternal.Reason, r)
	}
}

func TestFromError_NilReturnsNil(t *testing.T) {
	if FromError(nil) != nil {
		t.Error("FromError(nil) should return nil")
	}
}

func TestFromError_ErrorXPassthrough(t *testing.T) {
	original := ErrBind.WithMessage("bad json")
	result := FromError(original)
	if result.Code != original.Code || result.Reason != original.Reason {
		t.Error("FromError should pass through *ErrorX unchanged")
	}
}

func TestFromError_PlainError(t *testing.T) {
	result := FromError(errors.New("something broke"))
	if result.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", result.Code)
	}
	if result.Message != "something broke" {
		t.Errorf("expected message 'something broke', got %q", result.Message)
	}
}

func TestSentinelImmutability(t *testing.T) {
	origMsg := ErrInternal.Message
	origCode := ErrInternal.Code

	_ = ErrInternal.WithMessage("hacked")
	_ = ErrInternal.KV("foo", "bar")
	_ = ErrInternal.WithMetadata(map[string]string{"x": "y"})

	if ErrInternal.Message != origMsg {
		t.Errorf("sentinel message mutated: %q", ErrInternal.Message)
	}
	if ErrInternal.Code != origCode {
		t.Errorf("sentinel code mutated: %d", ErrInternal.Code)
	}
	if ErrInternal.Metadata != nil {
		t.Error("sentinel metadata mutated")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
