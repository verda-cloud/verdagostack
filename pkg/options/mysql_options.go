package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*MySQLOptions)(nil)

// MySQLOptions defines connection parameters for a MySQL database.
type MySQLOptions struct {
	Host                  string        `json:"host" mapstructure:"host"`
	Port                  int           `json:"port" mapstructure:"port"`
	Username              string        `json:"username" mapstructure:"username"`
	Password              string        `json:"password" mapstructure:"password"`
	Database              string        `json:"database" mapstructure:"database"`
	MaxIdleConnections    int           `json:"max-idle-connections" mapstructure:"max-idle-connections"`
	MaxOpenConnections    int           `json:"max-open-connections" mapstructure:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `json:"max-connection-life-time" mapstructure:"max-connection-life-time"`
	MaxConnectionIdleTime time.Duration `json:"max-connection-idle-time" mapstructure:"max-connection-idle-time"`
}

// NewMySQLOptions returns a MySQLOptions with production-ready defaults.
func NewMySQLOptions() *MySQLOptions {
	return &MySQLOptions{
		Host:                  "127.0.0.1",
		Port:                  3306,
		MaxIdleConnections:    10,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Hour,
		MaxConnectionIdleTime: 30 * time.Minute,
	}
}

// Validate verifies flags passed to MySQLOptions.
func (o *MySQLOptions) Validate() []error {
	var errs []error
	if o.Host == "" {
		errs = append(errs, fmt.Errorf("mysql host must not be empty"))
	}
	if o.Port < 1 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("mysql port must be between 1 and 65535"))
	}
	return errs
}

// AddFlags adds MySQL flags to the specified FlagSet.
func (o *MySQLOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Host, fullPrefix+".host", o.Host, "MySQL server host.")
	fs.IntVar(&o.Port, fullPrefix+".port", o.Port, "MySQL server port.")
	fs.StringVar(&o.Username, fullPrefix+".username", o.Username, "Username for MySQL authentication.")
	fs.StringVar(&o.Password, fullPrefix+".password", o.Password, "Password for MySQL authentication.")
	fs.StringVar(&o.Database, fullPrefix+".database", o.Database, "MySQL database name.")
	fs.IntVar(&o.MaxIdleConnections, fullPrefix+".max-idle-connections", o.MaxIdleConnections, "Maximum number of idle connections in the pool.")
	fs.IntVar(&o.MaxOpenConnections, fullPrefix+".max-open-connections", o.MaxOpenConnections, "Maximum number of open connections to the database.")
	fs.DurationVar(&o.MaxConnectionLifeTime, fullPrefix+".max-connection-life-time", o.MaxConnectionLifeTime, "Maximum lifetime of a connection.")
	fs.DurationVar(&o.MaxConnectionIdleTime, fullPrefix+".max-connection-idle-time", o.MaxConnectionIdleTime, "Maximum idle time of a connection.")
}

// DSN returns a MySQL connection string.
func (o *MySQLOptions) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		o.Username, o.Password, o.Host, o.Port, o.Database)
}
