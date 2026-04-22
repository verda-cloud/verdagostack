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

// ProviderOptions defines the service provider registration details for Polaris.
type ProviderOptions struct {
	Namespace string `json:"namespace" mapstructure:"namespace"`
	Service   string `json:"service" mapstructure:"service"`
	Host      string `json:"host" mapstructure:"host"`
	Port      int    `json:"port" mapstructure:"port"`
	Token     string `json:"token" mapstructure:"token"`
	Version   string `json:"version" mapstructure:"version"`
	Healthy   bool   `json:"healthy" mapstructure:"healthy"`
	TTL       int    `json:"ttl" mapstructure:"ttl"`
	Heartbeat bool   `json:"heartbeat" mapstructure:"heartbeat"`
}

// PolarisOptions defines options for the Polaris service discovery.
type PolarisOptions struct {
	Addr       string          `json:"addr" mapstructure:"addr"`
	Timeout    time.Duration   `json:"timeout" mapstructure:"timeout"`
	RetryCount int             `json:"retry-count" mapstructure:"retry-count"`
	Provider   ProviderOptions `json:"provider" mapstructure:"provider"`
}

// NewPolarisOptions creates a PolarisOptions object with default parameters.
func NewPolarisOptions() *PolarisOptions {
	return &PolarisOptions{
		Addr:       "127.0.0.1:8091",
		Timeout:    5 * time.Second,
		RetryCount: 3,
		Provider: ProviderOptions{
			Namespace: "default",
			Host:      GetLocalIP(),
			Version:   "v0.1.0",
			Healthy:   true,
			TTL:       1,
			Heartbeat: true,
		},
	}
}

// Validate verifies flags passed to PolarisOptions.
func (o *PolarisOptions) Validate() []error {
	return nil
}

// AddFlags adds Polaris flags to the specified FlagSet.
func (o *PolarisOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr, "Address of Polaris service (ip:port).")
	fs.DurationVar(&o.Timeout, fullPrefix+".timeout", o.Timeout, "Query timeout per request.")
	fs.IntVar(&o.RetryCount, fullPrefix+".retry-count", o.RetryCount, "Number of retries.")

	fs.StringVar(&o.Provider.Service, fullPrefix+".provider.service", o.Provider.Service, "Service name.")
	fs.StringVar(&o.Provider.Namespace, fullPrefix+".provider.namespace", o.Provider.Namespace, "Namespace.")
	fs.StringVar(&o.Provider.Host, fullPrefix+".provider.host", o.Provider.Host, "Service instance host (IPv4/IPv6).")
	fs.IntVar(&o.Provider.Port, fullPrefix+".provider.port", o.Provider.Port, "Service instance port.")
	fs.StringVar(&o.Provider.Token, fullPrefix+".provider.token", o.Provider.Token, "Service access token.")
	fs.BoolVar(&o.Provider.Heartbeat, fullPrefix+".provider.heartbeat", o.Provider.Heartbeat, "Let SDK handle heartbeat automatically.")
	fs.StringVar(&o.Provider.Version, fullPrefix+".provider.version", o.Provider.Version, "Service version.")
	fs.BoolVar(&o.Provider.Healthy, fullPrefix+".provider.healthy", o.Provider.Healthy, "Whether service is healthy.")
	fs.IntVar(&o.Provider.TTL, fullPrefix+".provider.ttl", o.Provider.TTL,
		"TTL timeout in seconds. Required if heartbeat is enabled.")
}
