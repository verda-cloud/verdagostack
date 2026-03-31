package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	genericoptions "github.com/verda-cloud/verdagostack/pkg/options"
)

// GRPCGatewayServer represents a gRPC-Gateway reverse proxy server.
type GRPCGatewayServer struct {
	srv *http.Server
}

// NewGRPCGatewayServer creates a new gRPC-Gateway reverse proxy server instance.
func NewGRPCGatewayServer(
	insecureOptions *genericoptions.InsecureServingOptions,
	grpcOptions *genericoptions.GRPCOptions,
	secureOptions *genericoptions.SecureServingOptions,
	registerHandler func(mux *runtime.ServeMux, conn *grpc.ClientConn) error,
) (*GRPCGatewayServer, error) {
	var tlsConfig *tls.Config
	if secureOptions != nil && secureOptions.Enabled {
		tlsConfig = secureOptions.MustTLSConfig()
		tlsConfig.InsecureSkipVerify = true //nolint:gosec
	}

	dialOptions := []grpc.DialOption{
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: 10 * time.Second,
		}),
	}
	if tlsConfig != nil {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(grpcOptions.Addr, dialOptions...)
	if err != nil {
		slog.Error("failed to create gRPC client connection", "error", err)
		return nil, err
	}

	gwmux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			// Output enum fields as numbers for consistency with protobuf definitions.
			UseEnumNumbers: true,
		},
	}))
	if err := registerHandler(gwmux, conn); err != nil {
		slog.Error("failed to register gRPC gateway handler", "error", err)
		return nil, err
	}

	return &GRPCGatewayServer{
		srv: &http.Server{
			Addr:              insecureOptions.Addr,
			Handler:           gwmux,
			TLSConfig:         tlsConfig,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

// Run starts the gRPC-Gateway server and blocks until it stops or fails.
func (s *GRPCGatewayServer) Run(ctx context.Context) error {
	slog.Info("starting gRPC-Gateway server", "protocol", protocolName(s.srv), "addr", s.srv.Addr)

	var err error
	if s.srv.TLSConfig != nil {
		err = s.srv.ListenAndServeTLS("", "")
	} else {
		err = s.srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to serve gRPC-Gateway server", "error", err)
		return err
	}
	return nil
}

// GracefulStop gracefully shuts down the gRPC-Gateway server.
func (s *GRPCGatewayServer) GracefulStop(ctx context.Context) error {
	slog.Info("gracefully stopping gRPC-Gateway server")
	if err := s.srv.Shutdown(ctx); err != nil {
		slog.Error("gRPC-Gateway server forced to shutdown", "error", err)
		return err
	}
	return nil
}
