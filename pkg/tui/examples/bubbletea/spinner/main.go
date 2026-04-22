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
	"time"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	_ "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea"
)

func main() {
	ctx := context.Background()
	s := tui.DefaultStatus()

	// Default spinner (dot style)
	spin, err := s.Spinner(ctx, "Cloning template repository...")
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	spin.UpdateMessage("Installing dependencies...")
	time.Sleep(1 * time.Second)
	spin.UpdateMessage("Generating configuration...")
	time.Sleep(1 * time.Second)
	spin.Stop("Template generated successfully")

	fmt.Println()

	// Globe spinner style
	spin, err = s.Spinner(ctx, "Fetching remote data...",
		tui.WithSpinnerStyle(tui.SpinnerGlobe),
	)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	spin.Stop("Data fetched from 3 regions")

	fmt.Println()

	// Moon spinner style
	spin, err = s.Spinner(ctx, "Running migrations...",
		tui.WithSpinnerStyle(tui.SpinnerMoon),
		tui.WithDoneSymbol("-->"),
	)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	spin.Stop("12 migrations applied")
}
