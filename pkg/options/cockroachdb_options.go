package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*CockroachDBOptions)(nil)

// CockroachDBOptions defines connection parameters for a CockroachDB database.
// CockroachDB uses the PostgreSQL wire protocol so the postgres GORM driver is used.
type CockroachDBOptions struct {
	Host                  string        `json:"host" mapstructure:"host"`
	Port                  int           `json:"port" mapstructure:"port"`
	Username              string        `json:"username" mapstructure:"username"`
	Password              string        `json:"password" mapstructure:"password"`
	Database              string        `json:"database" mapstructure:"database"`
	MaxIdleConnections    int           `json:"max-idle-connections" mapstructure:"max-idle-connections"`
	MaxOpenConnections    int           `json:"max-open-connections" mapstructure:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `json:"max-connection-life-time" mapstructure:"max-connection-life-time"`
	MaxConnectionIdleTime time.Duration `json:"max-connection-idle-time" mapstructure:"max-connection-idle-time"`
	SSLMode               string        `json:"sslmode" mapstructure:"sslmode"`
	Timezone              string        `json:"timezone" mapstructure:"timezone"`
}

// NewCockroachDBOptions returns a CockroachDBOptions with production-ready defaults.
func NewCockroachDBOptions() *CockroachDBOptions {
	return &CockroachDBOptions{
		Host:                  "127.0.0.1",
		Port:                  26257,
		MaxIdleConnections:    10,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Hour,
		MaxConnectionIdleTime: 30 * time.Minute,
		SSLMode:               "disable",
		Timezone:              "UTC",
	}
}

// Validate verifies flags passed to CockroachDBOptions.
func (o *CockroachDBOptions) Validate() []error {
	var errs []error
	if o.Host == "" {
		errs = append(errs, fmt.Errorf("cockroachdb host must not be empty"))
	}
	if o.Port < 1 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("cockroachdb port must be between 1 and 65535"))
	}
	return errs
}

// AddFlags adds CockroachDB flags to the specified FlagSet.
func (o *CockroachDBOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Host, fullPrefix+".host", o.Host, "CockroachDB server host.")
	fs.IntVar(&o.Port, fullPrefix+".port", o.Port, "CockroachDB server port.")
	fs.StringVar(&o.Username, fullPrefix+".username", o.Username, "Username for CockroachDB authentication.")
	fs.StringVar(&o.Password, fullPrefix+".password", o.Password, "Password for CockroachDB authentication.")
	fs.StringVar(&o.Database, fullPrefix+".database", o.Database, "CockroachDB database name.")
	fs.IntVar(&o.MaxIdleConnections, fullPrefix+".max-idle-connections", o.MaxIdleConnections, "Maximum number of idle connections in the pool.")
	fs.IntVar(&o.MaxOpenConnections, fullPrefix+".max-open-connections", o.MaxOpenConnections, "Maximum number of open connections to the database.")
	fs.DurationVar(&o.MaxConnectionLifeTime, fullPrefix+".max-connection-life-time", o.MaxConnectionLifeTime, "Maximum lifetime of a connection.")
	fs.DurationVar(&o.MaxConnectionIdleTime, fullPrefix+".max-connection-idle-time", o.MaxConnectionIdleTime, "Maximum idle time of a connection.")
	fs.StringVar(&o.SSLMode, fullPrefix+".sslmode", o.SSLMode, "CockroachDB SSL mode (disable, require, verify-ca, verify-full).")
	fs.StringVar(&o.Timezone, fullPrefix+".timezone", o.Timezone, "Session timezone for the CockroachDB connection.")
}

// DSN returns a CockroachDB-compatible PostgreSQL connection string.
func (o *CockroachDBOptions) DSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		o.Username, o.Password, o.Host, o.Port, o.Database, o.SSLMode)
}
