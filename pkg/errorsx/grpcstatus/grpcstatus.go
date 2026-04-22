// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package grpcstatus bridges errorsx.ErrorX with gRPC status codes. Import this
// package only in gRPC services; pure-HTTP or CLI code can use pkg/errorsx
// directly without pulling in gRPC dependencies.
package grpcstatus

import (
	"errors"
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/verda-cloud/verdagostack/pkg/errorsx"
)

// ToGRPCStatus converts an *ErrorX into a gRPC Status with ErrorInfo details.
func ToGRPCStatus(err *errorsx.ErrorX) *status.Status {
	details := errdetails.ErrorInfo{Reason: err.Reason, Metadata: err.Metadata}
	s, _ := status.New(httpToGRPC(err.Code), err.Message).WithDetails(&details)
	return s
}

// FromError converts any error into an *ErrorX, with full gRPC awareness.
// It checks (in order):
//  1. Whether err is already an *ErrorX
//  2. Whether err is a gRPC status error (extracts code, message, details)
//  3. Falls back to wrapping err as an internal server error
func FromError(err error) *errorsx.ErrorX {
	if err == nil {
		return nil
	}

	if errx := new(errorsx.ErrorX); errors.As(err, &errx) {
		return errx
	}

	gs, ok := status.FromError(err)
	if !ok {
		return errorsx.New(http.StatusInternalServerError, errorsx.ErrInternal.Reason, "%s", err.Error())
	}

	ret := errorsx.New(grpcToHTTP(gs.Code()), errorsx.ErrInternal.Reason, "%s", gs.Message())

	for _, detail := range gs.Details() {
		if typed, ok := detail.(*errdetails.ErrorInfo); ok {
			ret.Reason = typed.Reason
			return ret.WithMetadata(typed.Metadata)
		}
	}

	return ret
}

// httpToGRPC maps HTTP status codes to gRPC codes.
func httpToGRPC(httpCode int) codes.Code {
	switch httpCode {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	default:
		return codes.Internal
	}
}

// grpcToHTTP maps gRPC codes to HTTP status codes.
func grpcToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
