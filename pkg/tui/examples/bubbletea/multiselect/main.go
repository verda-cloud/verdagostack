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

	features := []string{
		"Authentication",
		"Rate Limiting",
		"Caching",
		"Logging",
		"Metrics",
		"Tracing",
	}

	// Multi-select with defaults and constraints
	indices, err := p.MultiSelect(ctx, "Enable features:",
		features,
		tui.WithMultiSelectDefaults([]int{3, 4}), // Logging and Metrics pre-selected
		tui.WithMinSelections(1),
		tui.WithMaxSelections(4),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Enabled:")
	for _, i := range indices {
		fmt.Printf("  - %s\n", features[i])
	}
}
