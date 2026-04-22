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
	"github.com/spf13/pflag"
)

var _ IOptions = (*ConsulOptions)(nil)

// ConsulOptions defines options for a Consul client connection.
type ConsulOptions struct {
	Addr   string `json:"addr,omitempty" mapstructure:"addr"`
	Scheme string `json:"scheme,omitempty" mapstructure:"scheme"`
}

// NewConsulOptions creates a ConsulOptions object with default parameters.
func NewConsulOptions() *ConsulOptions {
	return &ConsulOptions{
		Addr:   "127.0.0.1:8500",
		Scheme: "http",
	}
}

// Validate verifies flags passed to ConsulOptions.
func (o *ConsulOptions) Validate() []error {
	return nil
}

// AddFlags adds Consul flags to the specified FlagSet.
func (o *ConsulOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr, "Address of the Consul server.")
	fs.StringVar(&o.Scheme, fullPrefix+".scheme", o.Scheme, "URI scheme for the Consul server.")
}
