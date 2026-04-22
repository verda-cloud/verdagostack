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

package errorsx

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorX is a structured error type carrying an HTTP status code, a business
// reason string, a human-readable message, and optional metadata.
type ErrorX struct {
	Code     int               `json:"code,omitempty"`
	Reason   string            `json:"reason,omitempty"`
	Message  string            `json:"message,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// New creates a new ErrorX with the given code, reason, and formatted message.
func New(code int, reason string, format string, args ...any) *ErrorX {
	return &ErrorX{
		Code:    code,
		Reason:  reason,
		Message: fmt.Sprintf(format, args...),
	}
}

// Error implements the error interface.
func (err *ErrorX) Error() string {
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v", err.Code, err.Reason, err.Message, err.Metadata)
}

// WithMessage returns a copy of the error with an updated message.
func (err *ErrorX) WithMessage(format string, args ...any) *ErrorX {
	return &ErrorX{
		Code:     err.Code,
		Reason:   err.Reason,
		Message:  fmt.Sprintf(format, args...),
		Metadata: err.Metadata,
	}
}

// WithMetadata returns a copy of the error with the given metadata.
func (err *ErrorX) WithMetadata(md map[string]string) *ErrorX {
	return &ErrorX{
		Code:     err.Code,
		Reason:   err.Reason,
		Message:  err.Message,
		Metadata: md,
	}
}

// KV returns a copy of the error with additional key-value metadata pairs.
// Keys and values must be provided in alternating order.
func (err *ErrorX) KV(kvs ...string) *ErrorX {
	md := make(map[string]string, len(err.Metadata)+len(kvs)/2)
	for k, v := range err.Metadata {
		md[k] = v
	}
	for i := 0; i+1 < len(kvs); i += 2 {
		md[kvs[i]] = kvs[i+1]
	}
	return &ErrorX{
		Code:     err.Code,
		Reason:   err.Reason,
		Message:  err.Message,
		Metadata: md,
	}
}

// WithRequestID returns a copy of the error with the X-Request-ID metadata set.
func (err *ErrorX) WithRequestID(requestID string) *ErrorX {
	return err.KV("X-Request-ID", requestID)
}

// Is reports whether the error matches target by comparing Code and Reason.
func (err *ErrorX) Is(target error) bool {
	if errx := new(ErrorX); errors.As(target, &errx) {
		return errx.Code == err.Code && errx.Reason == err.Reason
	}
	return false
}

// Code returns the HTTP status code of err, or 200 if err is nil.
func Code(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return FromError(err).Code
}

// Reason returns the business reason string of err.
func Reason(err error) string {
	if err == nil {
		return ErrInternal.Reason
	}
	return FromError(err).Reason
}

// FromError converts any error into an *ErrorX. It checks whether err is
// already an *ErrorX and returns it directly; otherwise it wraps err as an
// internal server error. For gRPC-aware conversion (parsing gRPC status
// errors), use the grpcstatus sub-package.
func FromError(err error) *ErrorX {
	if err == nil {
		return nil
	}

	if errx := new(ErrorX); errors.As(err, &errx) {
		return errx
	}

	return New(ErrInternal.Code, ErrInternal.Reason, "%s", err.Error())
}
