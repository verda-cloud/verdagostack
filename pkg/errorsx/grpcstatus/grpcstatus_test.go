package grpcstatus

import (
	"errors"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/verda-cloud/verdagostack/pkg/errorsx"
)

func TestToGRPCStatus(t *testing.T) {
	err := errorsx.New(http.StatusNotFound, "NotFound", "user 42 not found")
	gs := ToGRPCStatus(err)
	if gs == nil {
		t.Fatal("ToGRPCStatus returned nil")
	}
	if gs.Code() != codes.NotFound {
		t.Errorf("expected gRPC NotFound, got %v", gs.Code())
	}
	if gs.Message() != "user 42 not found" {
		t.Errorf("expected message 'user 42 not found', got %q", gs.Message())
	}
}

func TestToGRPCStatus_WithMetadata(t *testing.T) {
	err := errorsx.ErrPermissionDenied.KV("resource", "secrets")
	gs := ToGRPCStatus(err)
	if gs.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", gs.Code())
	}
}

func TestFromError_Nil(t *testing.T) {
	if FromError(nil) != nil {
		t.Error("FromError(nil) should return nil")
	}
}

func TestFromError_ErrorX(t *testing.T) {
	original := errorsx.ErrBind.WithMessage("bad json")
	result := FromError(original)
	if result.Code != original.Code || result.Reason != original.Reason {
		t.Error("FromError should pass through *ErrorX unchanged")
	}
	if result.Message != "bad json" {
		t.Errorf("expected message 'bad json', got %q", result.Message)
	}
}

func TestFromError_PlainError(t *testing.T) {
	result := FromError(errors.New("disk full"))
	if result.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", result.Code)
	}
}

func TestFromError_GRPCStatus(t *testing.T) {
	gs := status.New(codes.NotFound, "item not found")
	result := FromError(gs.Err())
	if result.Code != http.StatusNotFound {
		t.Errorf("expected %d, got %d", http.StatusNotFound, result.Code)
	}
	if result.Message != "item not found" {
		t.Errorf("expected message 'item not found', got %q", result.Message)
	}
}

func TestRoundTrip_ErrorXToGRPCAndBack(t *testing.T) {
	original := errorsx.New(http.StatusForbidden, "Forbidden", "no access")
	gs := ToGRPCStatus(original)
	restored := FromError(gs.Err())

	if restored.Code != original.Code {
		t.Errorf("code: got %d, want %d", restored.Code, original.Code)
	}
	if restored.Reason != original.Reason {
		t.Errorf("reason: got %q, want %q", restored.Reason, original.Reason)
	}
	if restored.Message != original.Message {
		t.Errorf("message: got %q, want %q", restored.Message, original.Message)
	}
}

func TestHTTPToGRPC_Mapping(t *testing.T) {
	tests := []struct {
		http int
		grpc codes.Code
	}{
		{http.StatusOK, codes.OK},
		{http.StatusBadRequest, codes.InvalidArgument},
		{http.StatusUnauthorized, codes.Unauthenticated},
		{http.StatusForbidden, codes.PermissionDenied},
		{http.StatusNotFound, codes.NotFound},
		{http.StatusConflict, codes.AlreadyExists},
		{http.StatusTooManyRequests, codes.ResourceExhausted},
		{http.StatusNotImplemented, codes.Unimplemented},
		{http.StatusServiceUnavailable, codes.Unavailable},
		{http.StatusGatewayTimeout, codes.DeadlineExceeded},
		{http.StatusTeapot, codes.Internal},
	}
	for _, tc := range tests {
		if got := httpToGRPC(tc.http); got != tc.grpc {
			t.Errorf("httpToGRPC(%d) = %v, want %v", tc.http, got, tc.grpc)
		}
	}
}

func TestGRPCToHTTP_Mapping(t *testing.T) {
	tests := []struct {
		grpc codes.Code
		http int
	}{
		{codes.OK, http.StatusOK},
		{codes.InvalidArgument, http.StatusBadRequest},
		{codes.Unauthenticated, http.StatusUnauthorized},
		{codes.PermissionDenied, http.StatusForbidden},
		{codes.NotFound, http.StatusNotFound},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.Unimplemented, http.StatusNotImplemented},
		{codes.Unavailable, http.StatusServiceUnavailable},
		{codes.DeadlineExceeded, http.StatusGatewayTimeout},
		{codes.Internal, http.StatusInternalServerError},
	}
	for _, tc := range tests {
		if got := grpcToHTTP(tc.grpc); got != tc.http {
			t.Errorf("grpcToHTTP(%v) = %d, want %d", tc.grpc, got, tc.http)
		}
	}
}
