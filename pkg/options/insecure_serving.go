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
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*InsecureServingOptions)(nil)

// InsecureServingOptions contains configuration items related to HTTP server startup.
type InsecureServingOptions struct {
	Addr    string        `json:"addr" mapstructure:"addr"`
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewInsecureServingOptions creates an InsecureServingOptions object with default parameters.
func NewInsecureServingOptions() *InsecureServingOptions {
	return &InsecureServingOptions{
		Addr:    ":8080",
		Timeout: 30 * time.Second,
	}
}

// Validate verifies the flags passed to InsecureServingOptions.
func (o *InsecureServingOptions) Validate() []error {
	if o == nil {
		return nil
	}

	var errors []error

	if err := ValidateAddress(o.Addr); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// AddFlags registers HTTP server related flags to the specified FlagSet.
func (o *InsecureServingOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr,
		"Listen address for the HTTP server (e.g., :8080, 0.0.0.0:8443).")
	fs.DurationVar(&o.Timeout, fullPrefix+".timeout", o.Timeout,
		"Timeout for incoming HTTP connections.")
}

// Complete fills in any fields not set that are required to have valid data.
func (o *InsecureServingOptions) Complete() error {
	return nil
}
