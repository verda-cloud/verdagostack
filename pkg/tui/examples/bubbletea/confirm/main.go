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

	// Basic confirm with default=false (user must type y/Y to confirm)
	ok, err := p.Confirm(ctx, "Deploy to production?", tui.WithConfirmDefault(false))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Confirmed: %v\n", ok)

	// Confirm with default=true (pressing Enter confirms)
	ok, err = p.Confirm(ctx, "Continue with defaults?", tui.WithConfirmDefault(true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Continue: %v\n", ok)
}
