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

package gorm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/verda-cloud/verdagostack/pkg/log/empty"
	gormlogger "gorm.io/gorm/logger"
)

func TestGormLogger_ImplementsInterface(t *testing.T) {
	l := New(empty.NewLogger())
	var _ gormlogger.Interface = l
}

func TestGormLogger_LogMode(t *testing.T) {
	l := New(empty.NewLogger())

	silent := l.LogMode(gormlogger.Silent)
	if silent.(*GormLogger).level != gormlogger.Silent {
		t.Error("LogMode(Silent) did not set level")
	}

	warn := l.LogMode(gormlogger.Warn)
	if warn.(*GormLogger).level != gormlogger.Warn {
		t.Error("LogMode(Warn) did not set level")
	}

	// Original should not be mutated
	if l.level != gormlogger.Info {
		t.Error("LogMode mutated the original logger")
	}
}

func TestGormLogger_Info_Warn_Error_NoPanic(t *testing.T) {
	l := New(empty.NewLogger())
	ctx := context.Background()

	l.Info(ctx, "info msg %s", "test")
	l.Warn(ctx, "warn msg %s", "test")
	l.Error(ctx, "error msg %s", "test")
}

func TestGormLogger_Info_RespectedByLevel(t *testing.T) {
	l := New(empty.NewLogger())

	// At Silent level, nothing should panic
	silent := l.LogMode(gormlogger.Silent).(*GormLogger)
	silent.Info(context.Background(), "should be ignored")
	silent.Warn(context.Background(), "should be ignored")
	silent.Error(context.Background(), "should be ignored")
}

func TestGormLogger_Trace_NoPanic(t *testing.T) {
	l := New(empty.NewLogger())
	ctx := context.Background()
	begin := time.Now().Add(-100 * time.Millisecond)

	fc := func() (string, int64) {
		return "SELECT * FROM users", 10
	}

	// Normal trace
	l.Trace(ctx, begin, fc, nil)

	// Trace with error
	l.Trace(ctx, begin, fc, errors.New("query failed"))

	// Trace with slow query
	slowBegin := time.Now().Add(-500 * time.Millisecond)
	l.Trace(ctx, slowBegin, fc, nil)
}

func TestGormLogger_Trace_Silent(t *testing.T) {
	l := New(empty.NewLogger()).LogMode(gormlogger.Silent).(*GormLogger)
	ctx := context.Background()

	// Should not panic even at Silent
	l.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT 1", 0
	}, nil)
}

func TestGormLogger_Trace_IgnoresRecordNotFound(t *testing.T) {
	l := New(empty.NewLogger())
	ctx := context.Background()

	// ErrRecordNotFound should not trigger error-level logging
	// (no panic is sufficient since we use empty logger)
	l.Trace(ctx, time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "SELECT * FROM users WHERE id = 999", 0
	}, gormlogger.ErrRecordNotFound)
}

func TestGormLogger_WithSlowThreshold(t *testing.T) {
	l := New(empty.NewLogger())
	custom := l.WithSlowThreshold(500 * time.Millisecond)

	if custom.slowThreshold != 500*time.Millisecond {
		t.Errorf("expected 500ms threshold, got %v", custom.slowThreshold)
	}
	// Original unchanged
	if l.slowThreshold != defaultSlowThreshold {
		t.Error("WithSlowThreshold mutated the original")
	}
}

func TestGormLogger_Trace_NegativeRows(t *testing.T) {
	l := New(empty.NewLogger())
	ctx := context.Background()

	// rows=-1 is a valid GORM scenario (e.g., exec without RowsAffected)
	l.Trace(ctx, time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "DELETE FROM sessions", -1
	}, nil)

	l.Trace(ctx, time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "DELETE FROM sessions", -1
	}, errors.New("delete failed"))
}
