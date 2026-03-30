package db

import (
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/options"
)

func TestNewMySQL_AcceptsOptions(t *testing.T) {
	_ = NewMySQL
	_ = options.NewMySQLOptions()
}
