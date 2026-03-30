package db

import (
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/options"
)

func TestNewPostgreSQL_AcceptsOptions(t *testing.T) {
	_ = NewPostgreSQL
	_ = options.NewPostgreSQLOptions()
}
