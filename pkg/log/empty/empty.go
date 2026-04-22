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

// Package empty provides a no-op Logger implementation for use in tests
// and anywhere logging output should be silently discarded.
package empty

import (
	"context"

	"github.com/verda-cloud/verdagostack/pkg/log"
)

// EmptyLogger is a Logger that silently discards all output.
type EmptyLogger struct{}

var _ log.Logger = (*EmptyLogger)(nil)

// NewLogger returns a no-op Logger.
func NewLogger() log.Logger {
	return &EmptyLogger{}
}

func (l *EmptyLogger) Debugf(_ string, _ ...any) {}
func (l *EmptyLogger) Infof(_ string, _ ...any)  {}
func (l *EmptyLogger) Warnf(_ string, _ ...any)  {}
func (l *EmptyLogger) Errorf(_ string, _ ...any) {}
func (l *EmptyLogger) Panicf(_ string, _ ...any) {}
func (l *EmptyLogger) Fatalf(_ string, _ ...any) {}

func (l *EmptyLogger) Debugw(_ string, _ ...any) {}
func (l *EmptyLogger) Infow(_ string, _ ...any)  {}
func (l *EmptyLogger) Warnw(_ string, _ ...any)  {}
func (l *EmptyLogger) Errorw(_ string, _ ...any) {}
func (l *EmptyLogger) Panicw(_ string, _ ...any) {}
func (l *EmptyLogger) Fatalw(_ string, _ ...any) {}

func (l *EmptyLogger) W(_ context.Context) log.Logger { return l }
func (l *EmptyLogger) With(_ ...any) log.Logger       { return l }
func (l *EmptyLogger) AddCallerSkip(_ int) log.Logger { return l }
func (l *EmptyLogger) Sync()                          {}
