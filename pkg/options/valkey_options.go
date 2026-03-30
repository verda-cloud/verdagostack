package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*ValkeyOptions)(nil)

// ValkeyOptions defines connection parameters for a Valkey (or Redis-compatible) server.
type ValkeyOptions struct {
	Host       string         `json:"host" mapstructure:"host"`
	Port       int            `json:"port" mapstructure:"port"`
	Username   string         `json:"username" mapstructure:"username"`
	Password   string         `json:"password" mapstructure:"password"`
	Database   int            `json:"database" mapstructure:"database"`
	ClientName string         `json:"client-name" mapstructure:"client-name"`
	TLS        ValkeyTLS      `json:"tls" mapstructure:"tls"`
	Sentinel   ValkeySentinel `json:"sentinel" mapstructure:"sentinel"`
	TTL        time.Duration  `json:"ttl" mapstructure:"ttl"`
}

// ValkeyTLS holds TLS certificate paths for mTLS connections.
type ValkeyTLS struct {
	CACertPath string `json:"ca-cert-path" mapstructure:"ca-cert-path"`
	CertPath   string `json:"cert-path" mapstructure:"cert-path"`
	KeyPath    string `json:"key-path" mapstructure:"key-path"`
}

// Enabled returns true when both client cert and key are configured.
func (t ValkeyTLS) Enabled() bool {
	return t.CertPath != "" && t.KeyPath != ""
}

// ValkeySentinel holds configuration for Valkey Sentinel mode.
type ValkeySentinel struct {
	MasterName string   `json:"master-name" mapstructure:"master-name"`
	Addrs      []string `json:"addrs" mapstructure:"addrs"`
}

// Enabled returns true when a master name is configured.
func (s ValkeySentinel) Enabled() bool {
	return s.MasterName != ""
}

// NewValkeyOptions returns a ValkeyOptions with sensible defaults for a single-node setup.
func NewValkeyOptions() *ValkeyOptions {
	return &ValkeyOptions{
		Host:       "127.0.0.1",
		Port:       6379,
		ClientName: "verdastack",
		TTL:        24 * time.Hour,
	}
}

// Validate verifies flags passed to ValkeyOptions.
func (o *ValkeyOptions) Validate() []error {
	var errs []error
	if o.Host == "" && !o.Sentinel.Enabled() {
		errs = append(errs, fmt.Errorf("valkey host must not be empty when sentinel is disabled"))
	}
	if o.Port < 1 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("valkey port must be between 1 and 65535"))
	}
	if o.Sentinel.Enabled() && len(o.Sentinel.Addrs) == 0 {
		errs = append(errs, fmt.Errorf("valkey sentinel addrs must not be empty when sentinel is enabled"))
	}
	return errs
}

// AddFlags adds Valkey flags to the specified FlagSet.
func (o *ValkeyOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Host, fullPrefix+".host", o.Host, "Valkey server host.")
	fs.IntVar(&o.Port, fullPrefix+".port", o.Port, "Valkey server port.")
	fs.StringVar(&o.Username, fullPrefix+".username", o.Username, "Username for Valkey ACL authentication.")
	fs.StringVar(&o.Password, fullPrefix+".password", o.Password, "Password for Valkey authentication.")
	fs.IntVar(&o.Database, fullPrefix+".database", o.Database, "Valkey database number.")
	fs.StringVar(&o.ClientName, fullPrefix+".client-name", o.ClientName, "Client name sent to Valkey via CLIENT SETNAME.")
	fs.DurationVar(&o.TTL, fullPrefix+".ttl", o.TTL, "Default TTL for cached values.")
	fs.StringVar(&o.TLS.CACertPath, fullPrefix+".tls.ca-cert-path", o.TLS.CACertPath, "Path to the CA certificate for TLS.")
	fs.StringVar(&o.TLS.CertPath, fullPrefix+".tls.cert-path", o.TLS.CertPath, "Path to the client certificate for mTLS.")
	fs.StringVar(&o.TLS.KeyPath, fullPrefix+".tls.key-path", o.TLS.KeyPath, "Path to the client key for mTLS.")
	fs.StringVar(&o.Sentinel.MasterName, fullPrefix+".sentinel.master-name", o.Sentinel.MasterName, "Sentinel master name.")
	fs.StringSliceVar(&o.Sentinel.Addrs, fullPrefix+".sentinel.addrs", o.Sentinel.Addrs, "Comma-separated list of sentinel host:port addresses.")
}

// Addr returns the host:port address string.
func (o *ValkeyOptions) Addr() string {
	return fmt.Sprintf("%s:%d", o.Host, o.Port)
}
