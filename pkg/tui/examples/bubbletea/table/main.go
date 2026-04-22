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

package main

import (
	"context"
	"log"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	_ "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea"
)

func main() {
	ctx := context.Background()
	s := tui.DefaultStatus()

	// Service status table
	err := s.Table(ctx,
		[]string{"Service", "Status", "Endpoint", "Uptime"},
		[][]string{
			{"apigateway", "healthy", "http://localhost:9090", "3d 14h"},
			{"apiserver", "healthy", "http://localhost:9091", "3d 14h"},
			{"postgres", "healthy", "localhost:5432", "5d 2h"},
			{"redis", "degraded", "localhost:6379", "1d 8h"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
