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
	"fmt"
	"log"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	_ "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea"
)

func main() {
	ctx := context.Background()
	p := tui.Default()

	// Basic text input with placeholder
	name, err := p.TextInput(ctx, "Service name:", tui.WithPlaceholder("my-service"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Name: %s\n", name)

	// Text input with default value
	region, err := p.TextInput(ctx, "Region:", tui.WithDefault("us-east-1"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Region: %s\n", region)

	// Text input with validation
	port, err := p.TextInput(ctx, "Port:",
		tui.WithPlaceholder("8080"),
		tui.WithValidation(func(s string) error {
			if s == "" {
				return fmt.Errorf("port is required")
			}
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Port: %s\n", port)
}
