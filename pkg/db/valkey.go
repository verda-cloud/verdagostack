package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/valkey-io/valkey-go"

	"github.com/verda-cloud/verdagostack/pkg/options"
)

// NewValkey creates a valkey.Client from the provided options.
func NewValkey(opts *options.ValkeyOptions) (valkey.Client, error) {
	var tlsCfg *tls.Config
	if opts.TLS.Enabled() {
		var err error
		tlsCfg, err = buildValkeyTLS(&opts.TLS, opts.Host)
		if err != nil {
			return nil, err
		}
	}

	if opts.Sentinel.Enabled() {
		return newValkeySentinel(opts, tlsCfg)
	}
	return newValkeySingle(opts, tlsCfg)
}

func newValkeySingle(opts *options.ValkeyOptions, tlsCfg *tls.Config) (valkey.Client, error) {
	clientOpt := valkey.ClientOption{
		InitAddress:  []string{opts.Addr()},
		SelectDB:     opts.Database,
		ClientName:   opts.ClientName,
		TLSConfig:    tlsCfg,
		DisableCache: false,
	}
	if opts.Username != "" {
		clientOpt.Username = opts.Username
	}
	if opts.Password != "" {
		clientOpt.Password = opts.Password
	}

	client, err := valkey.NewClient(clientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}

	if err := client.Do(context.Background(), client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("valkey ping failed: %w", err)
	}

	return client, nil
}

func newValkeySentinel(opts *options.ValkeyOptions, tlsCfg *tls.Config) (valkey.Client, error) {
	clientOpt := valkey.ClientOption{
		InitAddress: opts.Sentinel.Addrs,
		Sentinel: valkey.SentinelOption{
			MasterSet: opts.Sentinel.MasterName,
			TLSConfig: tlsCfg,
		},
		ClientName: opts.ClientName,
		TLSConfig:  tlsCfg,
	}
	if opts.Username != "" {
		clientOpt.Username = opts.Username
		clientOpt.Sentinel.Username = opts.Username
	}
	if opts.Password != "" {
		clientOpt.Password = opts.Password
		clientOpt.Sentinel.Password = opts.Password
	}

	client, err := valkey.NewClient(clientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey sentinel client: %w", err)
	}
	return client, nil
}

func buildValkeyTLS(t *options.ValkeyTLS, serverHost string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(t.CertPath, t.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load valkey client certificate: %w", err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if t.CACertPath != "" {
		caCert, err := os.ReadFile(t.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read valkey CA certificate: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse valkey CA certificate")
		}
		cfg.RootCAs = pool
	}

	if serverHost != "" && serverHost != "127.0.0.1" && serverHost != "::1" {
		cfg.ServerName = serverHost
	} else {
		cfg.ServerName = "localhost"
	}

	return cfg, nil
}
