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

// Package kratos provides a Kratos logger adapter that delegates to a
// verdagostack log.Logger instance.
package kratos

import (
	"fmt"
	"os"

	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/verda-cloud/verdagostack/pkg/log"
)

// KratosLogger implements github.com/go-kratos/kratos/v2/log.Logger
// by delegating to a verdagostack log.Logger.
type KratosLogger struct {
	logger log.Logger
}

var _ krtlog.Logger = (*KratosLogger)(nil)

// New creates a KratosLogger backed by the given Logger.
func New(logger log.Logger) *KratosLogger {
	return &KratosLogger{logger: logger.AddCallerSkip(1)}
}

// Log implements krtlog.Logger. It expects an even number of keyvals;
// odd-length slices are logged as a warning and otherwise dropped.
func (l *KratosLogger) Log(level krtlog.Level, keyvals ...any) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.logger.Warnf("kratos logger: keyvals must appear in pairs: %v", keyvals)
		return nil
	}

	switch level {
	case krtlog.LevelDebug:
		l.logger.Debugw("", keyvals...)
	case krtlog.LevelInfo:
		l.logger.Infow("", keyvals...)
	case krtlog.LevelWarn:
		l.logger.Warnw("", keyvals...)
	case krtlog.LevelError:
		l.logger.Errorw("", keyvals...)
	case krtlog.LevelFatal:
		l.logger.Errorw(fmt.Sprint(keyvals...))
		os.Exit(1)
	}

	return nil
}
