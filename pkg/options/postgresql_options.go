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
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*PostgreSQLOptions)(nil)

// PostgreSQLOptions defines connection parameters for a PostgreSQL database.
type PostgreSQLOptions struct {
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
}

// NewPostgreSQLOptions returns a PostgreSQLOptions with production-ready defaults.
func NewPostgreSQLOptions() *PostgreSQLOptions {
	return &PostgreSQLOptions{
		Host:                  "127.0.0.1",
		Port:                  5432,
		MaxIdleConnections:    10,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Hour,
		MaxConnectionIdleTime: 30 * time.Minute,
		SSLMode:               "disable",
	}
}

// Validate verifies flags passed to PostgreSQLOptions.
func (o *PostgreSQLOptions) Validate() []error {
	var errs []error
	if o.Host == "" {
		errs = append(errs, fmt.Errorf("postgresql host must not be empty"))
	}
	if o.Port < 1 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("postgresql port must be between 1 and 65535"))
	}
	return errs
}

// AddFlags adds PostgreSQL flags to the specified FlagSet.
func (o *PostgreSQLOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Host, fullPrefix+".host", o.Host, "PostgreSQL server host.")
	fs.IntVar(&o.Port, fullPrefix+".port", o.Port, "PostgreSQL server port.")
	fs.StringVar(&o.Username, fullPrefix+".username", o.Username, "Username for PostgreSQL authentication.")
	fs.StringVar(&o.Password, fullPrefix+".password", o.Password, "Password for PostgreSQL authentication.")
	fs.StringVar(&o.Database, fullPrefix+".database", o.Database, "PostgreSQL database name.")
	fs.IntVar(&o.MaxIdleConnections, fullPrefix+".max-idle-connections", o.MaxIdleConnections, "Maximum number of idle connections in the pool.")
	fs.IntVar(&o.MaxOpenConnections, fullPrefix+".max-open-connections", o.MaxOpenConnections, "Maximum number of open connections to the database.")
	fs.DurationVar(&o.MaxConnectionLifeTime, fullPrefix+".max-connection-life-time", o.MaxConnectionLifeTime, "Maximum lifetime of a connection.")
	fs.DurationVar(&o.MaxConnectionIdleTime, fullPrefix+".max-connection-idle-time", o.MaxConnectionIdleTime, "Maximum idle time of a connection.")
	fs.StringVar(&o.SSLMode, fullPrefix+".sslmode", o.SSLMode, "PostgreSQL SSL mode (disable, require, verify-ca, verify-full).")
}

// DSN returns a PostgreSQL connection string.
func (o *PostgreSQLOptions) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		o.Host, o.Port, o.Username, o.Password, o.Database, o.SSLMode)
}
