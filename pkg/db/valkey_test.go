package db

import (
	"testing"

	"github.com/verda-cloud/verdagostack/pkg/options"
)

func TestNewValkey_AcceptsOptions(t *testing.T) {
	_ = NewValkey
	_ = options.NewValkeyOptions()
}
