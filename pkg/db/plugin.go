package db

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/verda-cloud/verdagostack/pkg/log"
)

const (
	callbackBeforeName = "verdastack:before"
	callbackAfterName  = "verdastack:after"
	startTimeKey       = "_verdastack_start_time"
)

// TracePlugin is a GORM plugin that measures SQL execution time and logs it
// via the verdastack log package.
type TracePlugin struct{}

var _ gorm.Plugin = (*TracePlugin)(nil)

func (TracePlugin) Name() string { return "verdastack:trace" }

func (TracePlugin) Initialize(db *gorm.DB) error {
	_ = db.Callback().Create().Before("gorm:before_create").Register(callbackBeforeName, traceBefore)
	_ = db.Callback().Query().Before("gorm:query").Register(callbackBeforeName, traceBefore)
	_ = db.Callback().Delete().Before("gorm:before_delete").Register(callbackBeforeName, traceBefore)
	_ = db.Callback().Update().Before("gorm:setup_reflect_value").Register(callbackBeforeName, traceBefore)
	_ = db.Callback().Row().Before("gorm:row").Register(callbackBeforeName, traceBefore)
	_ = db.Callback().Raw().Before("gorm:raw").Register(callbackBeforeName, traceBefore)

	_ = db.Callback().Create().After("gorm:after_create").Register(callbackAfterName, traceAfter)
	_ = db.Callback().Query().After("gorm:after_query").Register(callbackAfterName, traceAfter)
	_ = db.Callback().Delete().After("gorm:after_delete").Register(callbackAfterName, traceAfter)
	_ = db.Callback().Update().After("gorm:after_update").Register(callbackAfterName, traceAfter)
	_ = db.Callback().Row().After("gorm:row").Register(callbackAfterName, traceAfter)
	_ = db.Callback().Raw().After("gorm:raw").Register(callbackAfterName, traceAfter)

	return nil
}

func traceBefore(db *gorm.DB) {
	db.InstanceSet(startTimeKey, time.Now())
}

func traceAfter(db *gorm.DB) {
	val, ok := db.InstanceGet(startTimeKey)
	if !ok {
		return
	}
	start, ok := val.(time.Time)
	if !ok {
		return
	}
	log.Debugw("sql trace", "elapsed_ms", fmt.Sprintf("%.3f", float64(time.Since(start).Nanoseconds())/1e6))
}

// MustRawDB extracts the underlying *sql.DB from a GORM instance, panicking on error.
func MustRawDB(db *gorm.DB) *sql.DB {
	raw, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("db: failed to get *sql.DB from gorm: %v", err))
	}
	return raw
}
