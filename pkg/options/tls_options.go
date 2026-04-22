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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var _ IOptions = (*TLSOptions)(nil)

// TLSOptions defines the configuration for secure communication.
// It adopts a "Smart Mode": CA, Cert, and Key fields can be either file paths
// or the actual content (PEM string/Base64) itself.
type TLSOptions struct {
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
	SkipVerify bool   `json:"skip-verify" mapstructure:"skip-verify"`
	CA         string `json:"ca" mapstructure:"ca"`
	Cert       string `json:"cert" mapstructure:"cert"`
	Key        string `json:"key" mapstructure:"key"`
}

// NewTLSOptions creates a zero-value instance of TLSOptions with defaults.
func NewTLSOptions() TLSOptions {
	return TLSOptions{
		Enabled: false,
	}
}

// Validate verifies the flags passed to TLSOptions.
func (o *TLSOptions) Validate() []error {
	var errs []error

	if !o.Enabled {
		return errs
	}

	if (o.Cert != "" && o.Key == "") || (o.Cert == "" && o.Key != "") {
		errs = append(errs, fmt.Errorf("both cert and key must be provided to enable mTLS"))
	}

	return errs
}

// AddFlags adds flags to the specified FlagSet.
func (o *TLSOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.BoolVar(&o.Enabled, fullPrefix+".enabled", o.Enabled, "Enable TLS transport.")
	fs.BoolVar(&o.SkipVerify, fullPrefix+".skip-verify", o.SkipVerify, "Insecurely skip TLS certificate verification.")
	fs.StringVar(&o.CA, fullPrefix+".ca", o.CA, "Path to CA file OR raw PEM data.")
	fs.StringVar(&o.Cert, fullPrefix+".cert", o.Cert, "Path to Cert file OR raw PEM data.")
	fs.StringVar(&o.Key, fullPrefix+".key", o.Key, "Path to Key file OR raw PEM data.")
}

// MustTLSConfig returns a tls.Config or falls back to a default insecure config on error.
func (o *TLSOptions) MustTLSConfig() *tls.Config {
	tlsConf, err := o.TLSConfig()
	if err != nil {
		return &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	return tlsConf
}

// TLSConfig generates the final tls.Config from these options.
func (o *TLSOptions) TLSConfig() (*tls.Config, error) {
	if !o.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: o.SkipVerify, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
	}

	var certBytes, keyBytes []byte
	var err error

	if o.Cert != "" {
		certBytes, err = loadResource(o.Cert)
		if err != nil {
			return nil, fmt.Errorf("failed to load cert: %w", err)
		}
	}

	if o.Key != "" {
		keyBytes, err = loadResource(o.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to load key: %w", err)
		}
	}

	if len(certBytes) > 0 && len(keyBytes) > 0 {
		cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse x509 key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if o.CA != "" {
		caBytes, err := loadResource(o.CA)
		if err != nil {
			return nil, fmt.Errorf("failed to load ca: %w", err)
		}

		if len(caBytes) > 0 {
			capool := x509.NewCertPool()
			if !capool.AppendCertsFromPEM(caBytes) {
				return nil, fmt.Errorf("failed to append ca certs from pem")
			}
			tlsConfig.ClientCAs = capool
			tlsConfig.RootCAs = capool

			if tlsConfig.ClientCAs != nil {
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			}
		}
	}

	return tlsConfig, nil
}

// Scheme returns the URL scheme based on the TLS configuration.
func (o *TLSOptions) Scheme() string {
	if o.Enabled {
		return "https"
	}
	return "http"
}

// loadResource intelligently identifies whether the input is a file path,
// raw PEM string, or Base64-encoded PEM string.
func loadResource(input string) ([]byte, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	if strings.Contains(input, "-----BEGIN") {
		return []byte(input), nil
	}

	if !strings.Contains(input, "\n") && len(input) < 1024 {
		if _, err := os.Stat(input); err == nil {
			return os.ReadFile(input) //nolint:gosec // G304: file path comes from TLS config, not arbitrary user input
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err == nil {
		if strings.Contains(string(decoded), "-----BEGIN") {
			return decoded, nil
		}
	}

	return nil, fmt.Errorf("input is neither a valid file path, nor raw PEM data, nor base64 encoded PEM")
}
