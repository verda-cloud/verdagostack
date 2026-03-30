package options

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestCockroachDBOptions_Defaults(t *testing.T) {
	opts := NewCockroachDBOptions()
	if opts.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", opts.Host)
	}
	if opts.Port != 26257 {
		t.Errorf("expected port 26257, got %d", opts.Port)
	}
	if opts.Timezone != "UTC" {
		t.Errorf("expected timezone UTC, got %s", opts.Timezone)
	}
	if opts.SSLMode != "disable" {
		t.Errorf("expected sslmode disable, got %s", opts.SSLMode)
	}
}

func TestCockroachDBOptions_DSN(t *testing.T) {
	opts := &CockroachDBOptions{
		Host:     "crdb.example.com",
		Port:     26257,
		Username: "root",
		Password: "pass",
		Database: "billing",
		SSLMode:  "disable",
	}
	want := "postgresql://root:pass@crdb.example.com:26257/billing?sslmode=disable"
	if got := opts.DSN(); got != want {
		t.Errorf("DSN mismatch:\n got: %s\nwant: %s", got, want)
	}
}

func TestCockroachDBOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *CockroachDBOptions
		wantErr bool
	}{
		{"valid defaults", NewCockroachDBOptions(), false},
		{"empty host", &CockroachDBOptions{Host: "", Port: 26257}, true},
		{"invalid port", &CockroachDBOptions{Host: "localhost", Port: 0}, true},
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

func TestCockroachDBOptions_ImplementsIOptions(t *testing.T) {
	var _ IOptions = (*CockroachDBOptions)(nil)
}

func TestCockroachDBOptions_AddFlags(t *testing.T) {
	opts := NewCockroachDBOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs, "db.cockroachdb")

	for _, name := range []string{
		"db.cockroachdb.host",
		"db.cockroachdb.port",
		"db.cockroachdb.username",
		"db.cockroachdb.password",
		"db.cockroachdb.database",
		"db.cockroachdb.timezone",
		"db.cockroachdb.sslmode",
	} {
		if f := fs.Lookup(name); f == nil {
			t.Errorf("expected flag %s to be registered", name)
		}
	}
}
