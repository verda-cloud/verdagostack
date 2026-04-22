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
