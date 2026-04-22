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

func TestMySQLOptions_Defaults(t *testing.T) {
	opts := NewMySQLOptions()
	if opts.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", opts.Host)
	}
	if opts.Port != 3306 {
		t.Errorf("expected port 3306, got %d", opts.Port)
	}
	if opts.MaxIdleConnections != 10 {
		t.Errorf("expected max-idle-connections 10, got %d", opts.MaxIdleConnections)
	}
}

func TestMySQLOptions_DSN(t *testing.T) {
	opts := &MySQLOptions{
		Host:     "mysql.example.com",
		Port:     3307,
		Username: "user",
		Password: "pw",
		Database: "app",
	}
	want := "user:pw@tcp(mysql.example.com:3307)/app?charset=utf8mb4&parseTime=True&loc=Local"
	if got := opts.DSN(); got != want {
		t.Errorf("DSN mismatch:\n got: %s\nwant: %s", got, want)
	}
}

func TestMySQLOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *MySQLOptions
		wantErr bool
	}{
		{"valid defaults", NewMySQLOptions(), false},
		{"empty host", &MySQLOptions{Host: "", Port: 3306}, true},
		{"invalid port", &MySQLOptions{Host: "localhost", Port: -1}, true},
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

func TestMySQLOptions_ImplementsIOptions(t *testing.T) {
	var _ IOptions = (*MySQLOptions)(nil)
}

func TestMySQLOptions_AddFlags(t *testing.T) {
	opts := NewMySQLOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs, "db.mysql")

	for _, name := range []string{
		"db.mysql.host",
		"db.mysql.port",
		"db.mysql.username",
		"db.mysql.password",
		"db.mysql.database",
		"db.mysql.max-idle-connections",
		"db.mysql.max-open-connections",
	} {
		if f := fs.Lookup(name); f == nil {
			t.Errorf("expected flag %s to be registered", name)
		}
	}
}
