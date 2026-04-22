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

var _ IOptions = (*EtcdOptions)(nil)

// EtcdOptions defines options for an etcd cluster connection.
type EtcdOptions struct {
	Endpoints   []string      `json:"endpoints" mapstructure:"endpoints"`
	DialTimeout time.Duration `json:"dial-timeout" mapstructure:"dial-timeout"`
	Username    string        `json:"username" mapstructure:"username"`
	Password    string        `json:"password" mapstructure:"password"`
	TLSOptions  TLSOptions    `json:"tls" mapstructure:"tls"`
}

// NewEtcdOptions creates an EtcdOptions object with default parameters.
func NewEtcdOptions() *EtcdOptions {
	return &EtcdOptions{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
		TLSOptions:  NewTLSOptions(),
	}
}

// Validate verifies flags passed to EtcdOptions.
func (o *EtcdOptions) Validate() []error {
	var errs []error

	if len(o.Endpoints) == 0 {
		errs = append(errs, fmt.Errorf("--etcd.endpoints can not be empty"))
	}

	if o.DialTimeout <= 0 {
		errs = append(errs, fmt.Errorf("--etcd.dial-timeout cannot be negative"))
	}

	errs = append(errs, o.TLSOptions.Validate()...)

	return errs
}

// AddFlags adds etcd flags to the specified FlagSet.
func (o *EtcdOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	o.TLSOptions.AddFlags(fs, fullPrefix+".tls")
	fs.StringSliceVar(&o.Endpoints, fullPrefix+".endpoints", o.Endpoints, "Endpoints of etcd cluster.")
	fs.StringVar(&o.Username, fullPrefix+".username", o.Username, "Username of etcd cluster.")
	fs.StringVar(&o.Password, fullPrefix+".password", o.Password, "Password of etcd cluster.")
	fs.DurationVar(&o.DialTimeout, fullPrefix+".dial-timeout", o.DialTimeout, "Etcd dial timeout in seconds.")
}
