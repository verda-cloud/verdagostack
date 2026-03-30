package server

import (
	"context"

	"github.com/gin-gonic/gin"

	genericoptions "github.com/verda-cloud/verdagostack/pkg/options"
)

// GinServer wraps a Gin engine as a Server implementation, supporting both
// HTTP and HTTPS. It composes an HTTPServer internally to reuse the
// dual-listener and graceful shutdown logic.
type GinServer struct {
	engine     *gin.Engine
	httpServer *HTTPServer
}

// NewGinServer creates a new GinServer with the given Gin engine and serving options.
// Pass nil for secureOptions to disable HTTPS.
func NewGinServer(
	insecureOptions *genericoptions.InsecureServingOptions,
	secureOptions *genericoptions.SecureServingOptions,
	engine *gin.Engine,
) *GinServer {
	return &GinServer{
		engine:     engine,
		httpServer: NewHTTPServer(insecureOptions, secureOptions, engine),
	}
}

// Engine returns the underlying Gin engine for route registration.
func (s *GinServer) Engine() *gin.Engine {
	return s.engine
}

// Run starts the Gin server and blocks until it stops or fails.
func (s *GinServer) Run(ctx context.Context) error {
	return s.httpServer.Run(ctx)
}

// GracefulStop gracefully shuts down the Gin server.
func (s *GinServer) GracefulStop(ctx context.Context) error {
	return s.httpServer.GracefulStop(ctx)
}
