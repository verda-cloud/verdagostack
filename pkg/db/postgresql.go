package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/verda-cloud/verdagostack/pkg/options"
)

// NewPostgreSQL creates a GORM DB instance connected to PostgreSQL.
// If logger is nil, the default GORM logger is used.
func NewPostgreSQL(opts *options.PostgreSQLOptions, logger gormlogger.Interface) (*gorm.DB, error) {
	if logger == nil {
		logger = gormlogger.Default
	}

	db, err := gorm.Open(postgres.Open(opts.DSN()), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)
	sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifeTime)
	sqlDB.SetConnMaxIdleTime(opts.MaxConnectionIdleTime)

	return db, nil
}
