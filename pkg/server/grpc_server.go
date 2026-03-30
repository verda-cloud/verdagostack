package server

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	genericoptions "github.com/verda-cloud/verdagostack/pkg/options"
)

// GRPCServer represents a gRPC server.
type GRPCServer struct {
	srv *grpc.Server
	lis net.Listener
}

// NewGRPCServer creates a new gRPC server instance.
// The registerBuilder function returns a service registration function and
// a service name used for health checks.
func NewGRPCServer(
	grpcOptions *genericoptions.GRPCOptions,
	tlsOptions *genericoptions.TLSOptions,
	serverOptions []grpc.ServerOption,
	registerBuilder func() (func(grpc.ServiceRegistrar), string),
) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", grpcOptions.Addr)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return nil, err
	}

	if tlsOptions != nil && tlsOptions.Enabled {
		tlsConfig := tlsOptions.MustTLSConfig()
		serverOptions = append(serverOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcsrv := grpc.NewServer(serverOptions...)

	registerFn, serverName := registerBuilder()
	registerFn(grpcsrv)
	registerHealthServer(serverName, grpcsrv)
	reflection.Register(grpcsrv)

	return &GRPCServer{
		srv: grpcsrv,
		lis: lis,
	}, nil
}

// Run starts the gRPC server and blocks until it stops or fails.
func (s *GRPCServer) Run(ctx context.Context) error {
	slog.Info("starting gRPC server", "addr", s.lis.Addr().String())
	if err := s.srv.Serve(s.lis); err != nil {
		slog.Error("failed to serve gRPC server", "error", err)
		return err
	}
	return nil
}

// GracefulStop gracefully shuts down the gRPC server.
func (s *GRPCServer) GracefulStop(ctx context.Context) error {
	slog.Info("gracefully stopping gRPC server")
	s.srv.GracefulStop()
	return nil
}

// registerHealthServer registers a gRPC health check service.
func registerHealthServer(serverName string, grpcsrv *grpc.Server) {
	healthServer := health.NewServer()
	healthServer.SetServingStatus(serverName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcsrv, healthServer)
}
