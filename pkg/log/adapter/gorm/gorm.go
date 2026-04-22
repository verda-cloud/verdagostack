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

// Package gorm provides a GORM logger adapter that delegates to a
// verdagostack log.Logger instance.
package gorm

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/verda-cloud/verdagostack/pkg/log"
	gormlogger "gorm.io/gorm/logger"
)

const defaultSlowThreshold = 200 * time.Millisecond

// GormLogger implements gorm.io/gorm/logger.Interface by delegating to a
// verdagostack log.Logger.
type GormLogger struct {
	logger        log.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

var _ gormlogger.Interface = (*GormLogger)(nil)

// New creates a GormLogger backed by the given Logger.
// The default log level is Info and slow-query threshold is 200ms.
func New(logger log.Logger) *GormLogger {
	return &GormLogger{
		logger:        logger.AddCallerSkip(1),
		level:         gormlogger.Info,
		slowThreshold: defaultSlowThreshold,
	}
}

// WithSlowThreshold returns a copy with the specified slow-query threshold.
func (l *GormLogger) WithSlowThreshold(d time.Duration) *GormLogger {
	c := *l
	c.slowThreshold = d
	return &c
}

// LogMode returns a copy configured to the requested GORM log level.
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	c := *l
	c.level = level
	return &c
}

func (l *GormLogger) Info(_ context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Info {
		l.logger.Infof(msg, args...)
	}
}

func (l *GormLogger) Warn(_ context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Warn {
		l.logger.Warnf(msg, args...)
	}
}

func (l *GormLogger) Error(_ context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Error {
		l.logger.Errorf(msg, args...)
	}
}

func (l *GormLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	caller := fileWithLineNum()

	switch {
	case err != nil && !errors.Is(err, gormlogger.ErrRecordNotFound) && l.level >= gormlogger.Error:
		l.logger.Errorw("SQL error",
			"caller", caller,
			"err", err,
			"elapsed_ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6),
			"rows", rows,
			"sql", sql,
		)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		l.logger.Warnw("slow SQL",
			"caller", caller,
			"elapsed_ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6),
			"threshold", l.slowThreshold.String(),
			"rows", rows,
			"sql", sql,
		)
	case l.level >= gormlogger.Info:
		l.logger.Infow("SQL",
			"caller", caller,
			"elapsed_ms", fmt.Sprintf("%.3f", float64(elapsed.Nanoseconds())/1e6),
			"rows", rows,
			"sql", sql,
		)
	}
}

func fileWithLineNum() string {
	for i := 4; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && !strings.Contains(file, "gorm.io") && !strings.HasSuffix(file, "_test.go") {
			dir, f := filepath.Split(file)
			return filepath.Join(filepath.Base(dir), f) + ":" + strconv.Itoa(line)
		}
	}
	return ""
}
