package db

import (
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/options"
)

func TestNewCockroachDB_AcceptsOptions(t *testing.T) {
	_ = NewCockroachDB
	_ = options.NewCockroachDBOptions()
}
