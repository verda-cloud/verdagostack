package options

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestValkeyOptions_Defaults(t *testing.T) {
	opts := NewValkeyOptions()
	if opts.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", opts.Host)
	}
	if opts.Port != 6379 {
		t.Errorf("expected port 6379, got %d", opts.Port)
	}
	if opts.ClientName != "verdagostack" {
		t.Errorf("expected client name verdagostack, got %s", opts.ClientName)
	}
}

func TestValkeyOptions_Addr(t *testing.T) {
	opts := &ValkeyOptions{Host: "cache.example.com", Port: 6380}
	want := "cache.example.com:6380"
	if got := opts.Addr(); got != want {
		t.Errorf("Addr() = %s, want %s", got, want)
	}
}

func TestValkeyOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *ValkeyOptions
		wantErr bool
	}{
		{"valid defaults", NewValkeyOptions(), false},
		{"empty host no sentinel", &ValkeyOptions{Host: "", Port: 6379}, true},
		{"invalid port", &ValkeyOptions{Host: "localhost", Port: 0}, true},
		{
			"sentinel enabled without addrs",
			&ValkeyOptions{
				Host: "localhost",
				Port: 6379,
				Sentinel: ValkeySentinel{
					MasterName: "mymaster",
					Addrs:      nil,
				},
			},
			true,
		},
		{
			"sentinel enabled with addrs",
			&ValkeyOptions{
				Host: "",
				Port: 6379,
				Sentinel: ValkeySentinel{
					MasterName: "mymaster",
					Addrs:      []string{"sentinel1:26379", "sentinel2:26379"},
				},
			},
			false,
		},
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

func TestValkeyOptions_ImplementsIOptions(t *testing.T) {
	var _ IOptions = (*ValkeyOptions)(nil)
}

func TestValkeyOptions_AddFlags(t *testing.T) {
	opts := NewValkeyOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs, "cache.valkey")

	for _, name := range []string{
		"cache.valkey.host",
		"cache.valkey.port",
		"cache.valkey.username",
		"cache.valkey.password",
		"cache.valkey.database",
		"cache.valkey.client-name",
		"cache.valkey.ttl",
		"cache.valkey.tls.ca-cert-path",
		"cache.valkey.tls.cert-path",
		"cache.valkey.tls.key-path",
		"cache.valkey.sentinel.master-name",
		"cache.valkey.sentinel.addrs",
	} {
		if f := fs.Lookup(name); f == nil {
			t.Errorf("expected flag %s to be registered", name)
		}
	}
}

func TestValkeyTLS_Enabled(t *testing.T) {
	tests := []struct {
		name string
		tls  ValkeyTLS
		want bool
	}{
		{"empty", ValkeyTLS{}, false},
		{"only cert", ValkeyTLS{CertPath: "/cert.pem"}, false},
		{"only key", ValkeyTLS{KeyPath: "/key.pem"}, false},
		{"both", ValkeyTLS{CertPath: "/cert.pem", KeyPath: "/key.pem"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tls.Enabled(); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValkeySentinel_Enabled(t *testing.T) {
	if (ValkeySentinel{}).Enabled() {
		t.Error("empty sentinel should not be enabled")
	}
	if !(ValkeySentinel{MasterName: "mymaster"}).Enabled() {
		t.Error("sentinel with master name should be enabled")
	}
}
