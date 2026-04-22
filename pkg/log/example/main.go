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

// Command example demonstrates all major features of the verdagostack pkg/log package.
//
// Run with:
//
//	go run ./pkg/log/example
//	go run ./pkg/log/example --log.level=debug --log.format=json
//	go run ./pkg/log/example --log.enable-color
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/verda-cloud/verdagostack/pkg/log"
	logadapterkratos "github.com/verda-cloud/verdagostack/pkg/log/adapter/kratos"
	"github.com/verda-cloud/verdagostack/pkg/log/empty"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

func main() {
	// ----------------------------------------------------------------
	// 1. Options + pflag CLI integration
	// ----------------------------------------------------------------
	opts := log.NewOptions()
	opts.AddFlags(pflag.CommandLine)
	pflag.Parse()

	log.Init(opts, log.WithContextExtractor(log.ContextExtractors{
		"request.id": func(ctx context.Context) string {
			if v, ok := ctx.Value(requestIDKey).(string); ok {
				return v
			}
			return ""
		},
	}))
	defer log.Sync()

	fmt.Println("=== 1. Package-level convenience functions ===")
	log.Infow("server starting", "port", 8080, "env", "production")
	log.Infof("listening on :%d", 8080)
	log.Warnw("deprecated config key", "key", "old_db_url")
	log.Debugw("this only shows at debug level", "detail", "hidden by default")

	// ----------------------------------------------------------------
	// 2. Dependency injection via Logger interface
	// ----------------------------------------------------------------
	fmt.Println("\n=== 2. Dependency injection ===")
	svc := NewGreetService(log.Default())
	svc.Greet("Alice")

	// ----------------------------------------------------------------
	// 3. With() — child logger with pre-set fields
	// ----------------------------------------------------------------
	fmt.Println("\n=== 3. Child logger via With() ===")
	componentLog := log.With("component", "auth", "version", "v2")
	componentLog.Infow("module initialized")
	componentLog.Warnw("token expiring soon", "ttl_seconds", 30)

	// ----------------------------------------------------------------
	// 4. W(ctx) — context-aware logging
	// ----------------------------------------------------------------
	fmt.Println("\n=== 4. Context-aware logging via W(ctx) ===")
	ctx := context.WithValue(context.Background(), requestIDKey, "req-abc-123")
	log.W(ctx).Infow("handling request", "method", "GET", "path", "/api/users")
	log.W(ctx).Errorf("something went wrong: %s", "timeout")

	// Without a request ID in context — field is omitted
	log.W(context.Background()).Infow("background task", "task", "cleanup")

	// ----------------------------------------------------------------
	// 5. Error logging with structured error field
	// ----------------------------------------------------------------
	fmt.Println("\n=== 5. Error logging ===")
	err := errors.New("connection refused")
	log.Errorw("database connection failed", "err", err, "host", "db.internal", "port", 5432)

	// ----------------------------------------------------------------
	// 6. Separate Logger instance (non-global)
	// ----------------------------------------------------------------
	fmt.Println("\n=== 6. Standalone logger instance ===")
	jsonOpts := &log.Options{
		Level:       "debug",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}
	jsonLogger := log.NewLogger(jsonOpts)
	jsonLogger.Debugw("json logger", "format", "json", "standalone", true)

	// ----------------------------------------------------------------
	// 7. Kratos adapter
	// ----------------------------------------------------------------
	fmt.Println("\n=== 7. Kratos adapter ===")
	kratosLog := logadapterkratos.New(log.Default())
	_ = kratosLog.Log(1, "msg", "kratos info message", "service", "gateway") // krtlog.LevelInfo = 1

	// ----------------------------------------------------------------
	// 8. Empty logger for testing
	// ----------------------------------------------------------------
	fmt.Println("\n=== 8. Empty logger (no output expected) ===")
	silent := empty.NewLogger()
	silent.Infow("this should produce no output")
	silent.Errorf("this too: %s", "nothing")
	fmt.Println("(silence is golden)")

	// ----------------------------------------------------------------
	// 9. Multi-output paths
	// ----------------------------------------------------------------
	fmt.Println("\n=== 9. Multi-output paths ===")
	tmpFile := "/tmp/verdagostack-log-example.log"
	multiOpts := &log.Options{
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout", tmpFile},
	}
	multiLogger := log.NewLogger(multiOpts)
	multiLogger.Infow("written to both stdout and file", "file", tmpFile)
	multiLogger.Sync()

	if data, err := os.ReadFile(tmpFile); err == nil {
		fmt.Printf("File contents: %s", data)
		_ = os.Remove(tmpFile)
	}

	fmt.Println("\n=== Done ===")
}

// GreetService demonstrates injecting a Logger via constructor.
type GreetService struct {
	log log.Logger
}

func NewGreetService(logger log.Logger) *GreetService {
	return &GreetService{log: logger.With("service", "greet")}
}

func (s *GreetService) Greet(name string) {
	s.log.Infow("greeting user", "name", name)
	s.log.Infof("hello, %s!", name)
}
