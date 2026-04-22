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

package server

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v2"
	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/verda-cloud/verdagostack/pkg/log"
	kratosadapter "github.com/verda-cloud/verdagostack/pkg/log/adapter/kratos"
)

// KratosAppConfig holds the configuration for creating a Kratos application.
type KratosAppConfig struct {
	ID        string
	Name      string
	Version   string
	Metadata  map[string]string
	Registrar registry.Registrar
}

// KratosServer wraps a Kratos application as a Server implementation.
type KratosServer struct {
	kapp *kratos.App
}

// NewKratosServer creates a new KratosServer with the given configuration and transport servers.
func NewKratosServer(cfg KratosAppConfig, logger log.Logger, servers ...transport.Server) (*KratosServer, error) {
	kapp := kratos.New(
		kratos.ID(cfg.ID+"."+cfg.Name),
		kratos.Name(cfg.Name),
		kratos.Version(cfg.Version),
		kratos.Metadata(cfg.Metadata),
		kratos.Logger(NewKratosLogger(logger, cfg.ID, cfg.Name, cfg.Version)),
		kratos.Registrar(cfg.Registrar),
		kratos.Server(servers...),
	)

	return &KratosServer{kapp: kapp}, nil
}

// Run starts the Kratos application and blocks until it stops or fails.
func (s *KratosServer) Run(ctx context.Context) error {
	slog.Info("starting Kratos application")
	if err := s.kapp.Run(); err != nil {
		slog.Error("failed to run Kratos application", "error", err)
		return err
	}
	return nil
}

// GracefulStop gracefully shuts down the Kratos application.
func (s *KratosServer) GracefulStop(ctx context.Context) error {
	slog.Info("gracefully stopping Kratos application")
	if err := s.kapp.Stop(); err != nil {
		slog.Error("failed to gracefully stop Kratos application", "error", err)
		return err
	}
	return nil
}

// NewKratosLogger creates a Kratos-compatible logger backed by the given verdagostack Logger,
// enriched with service metadata fields.
func NewKratosLogger(logger log.Logger, id, name, version string) krtlog.Logger {
	return krtlog.With(kratosadapter.New(logger),
		"ts", krtlog.DefaultTimestamp,
		"caller", krtlog.DefaultCaller,
		"service.id", id,
		"service.name", name,
		"service.version", version,
	)
}
