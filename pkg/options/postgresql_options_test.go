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

package options

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestPostgreSQLOptions_Defaults(t *testing.T) {
	opts := NewPostgreSQLOptions()
	if opts.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", opts.Host)
	}
	if opts.Port != 5432 {
		t.Errorf("expected port 5432, got %d", opts.Port)
	}
	if opts.SSLMode != "disable" {
		t.Errorf("expected sslmode disable, got %s", opts.SSLMode)
	}
	if opts.MaxIdleConnections != 10 {
		t.Errorf("expected max-idle-connections 10, got %d", opts.MaxIdleConnections)
	}
	if opts.MaxOpenConnections != 100 {
		t.Errorf("expected max-open-connections 100, got %d", opts.MaxOpenConnections)
	}
}

func TestPostgreSQLOptions_DSN(t *testing.T) {
	opts := &PostgreSQLOptions{
		Host:     "db.example.com",
		Port:     5433,
		Username: "admin",
		Password: "secret",
		Database: "mydb",
		SSLMode:  "require",
	}
	want := "host=db.example.com port=5433 user=admin password=secret dbname=mydb sslmode=require"
	if got := opts.DSN(); got != want {
		t.Errorf("DSN mismatch:\n got: %s\nwant: %s", got, want)
	}
}

func TestPostgreSQLOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *PostgreSQLOptions
		wantErr bool
	}{
		{"valid defaults", NewPostgreSQLOptions(), false},
		{"empty host", &PostgreSQLOptions{Host: "", Port: 5432}, true},
		{"invalid port zero", &PostgreSQLOptions{Host: "localhost", Port: 0}, true},
		{"invalid port high", &PostgreSQLOptions{Host: "localhost", Port: 70000}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.opts.Validate()
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors, got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected validation errors: %v", errs)
			}
		})
	}
}

func TestPostgreSQLOptions_ImplementsIOptions(t *testing.T) {
	var _ IOptions = (*PostgreSQLOptions)(nil)
}

func TestPostgreSQLOptions_AddFlags(t *testing.T) {
	opts := NewPostgreSQLOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs, "db.postgresql")

	for _, name := range []string{
		"db.postgresql.host",
		"db.postgresql.port",
		"db.postgresql.username",
		"db.postgresql.password",
		"db.postgresql.database",
		"db.postgresql.max-idle-connections",
		"db.postgresql.max-open-connections",
		"db.postgresql.sslmode",
	} {
		if f := fs.Lookup(name); f == nil {
			t.Errorf("expected flag %s to be registered", name)
		}
	}
}
