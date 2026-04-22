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

	environments := []string{"development", "staging", "production"}

	// Basic select
	idx, err := p.Select(ctx, "Target environment:", environments)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected: %s (index %d)\n", environments[idx], idx)

	// Select with default and page size
	regions := []string{
		"us-east-1", "us-west-2", "eu-west-1", "eu-central-1",
		"ap-southeast-1", "ap-northeast-1", "sa-east-1",
	}
	idx, err = p.Select(ctx, "Region:",
		regions,
		tui.WithSelectDefault(1),
		tui.WithPageSize(5),
		tui.WithLoop(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Region: %s\n", regions[idx])
}
