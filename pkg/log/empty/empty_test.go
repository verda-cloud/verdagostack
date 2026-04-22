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

package empty

import (
	"context"
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/log"
)

func TestEmptyLogger_ImplementsInterface(t *testing.T) {
	var logger log.Logger = NewLogger() //nolint:staticcheck // explicit interface check
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
