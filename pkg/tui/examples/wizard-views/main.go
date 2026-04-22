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

// Command wizard-views demonstrates the wizard engine's View and pub/sub
// message bus. It shows how Views react to engine events and communicate
// with each other through typed messages.
//
// Run:
//
//	go run ./pkg/tui/examples/wizard-views
//
// What you'll see:
//
//	████████████████░░░░░░░░░░  Step 2 of 4    <- ProgressView (built-in)
//	  Estimated cost: $0.35/hr                  <- CostView (custom, reacts to CollectedChanged)
//	  Quota: 8 of 10 instances remaining        <- QuotaView (custom, reacts to CostView's message)
//	? Select instance size                      <- Prompt (engine's sequential loop)
//	  > Small (2 vCPU, 4GB)
//	    Medium (4 vCPU, 8GB)
//	    Large (8 vCPU, 16GB)
//
// This example demonstrates three concepts:
//
//  1. Engine → View: The engine broadcasts StepChangedMsg and CollectedChangedMsg.
//     Views subscribe and render based on the data in those messages.
//
//  2. View → View (pub/sub by type): CostView publishes CostUpdatedMsg when it
//     recalculates. QuotaView subscribes to CostUpdatedMsg and renders remaining
//     quota. They don't know about each other — they only know the message type.
//
//  3. Route by type: The message bus looks at the Go struct type of each message
//     (e.g., CostUpdatedMsg) and delivers it to views that declared that type in
//     their Subscribe() method. No names, no channels — just type matching.
package main

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	"github.com/verda-cloud/verdagostack/pkg/tui/wizard"

	"charm.land/lipgloss/v2"
)

// --- Pricing data ---

var prices = map[string]map[string]float64{
	"us-east": {"small": 0.10, "medium": 0.35, "large": 0.80},
	"eu-west": {"small": 0.12, "medium": 0.40, "large": 0.90},
}

var quotas = map[string]int{
	"us-east": 10,
	"eu-west": 5,
}

// --- Custom message types ---
// These are the "topics" that views publish and subscribe to.
// The bus routes by these Go struct types.

// CostUpdatedMsg is published by CostView when it recalculates the price.
// Any view can subscribe to this type to react to price changes.
type CostUpdatedMsg struct {
	Region string
	Size   string
	Price  float64
}

// --- CostView: reacts to engine's CollectedChangedMsg ---

// CostView displays estimated hourly cost. It listens to CollectedChangedMsg
// (broadcast by the engine after each step completes) and recalculates the
// price when both "region" and "size" are collected.
//
// It also PUBLISHES CostUpdatedMsg so other views can react to price changes.
type CostView struct {
	last   string
	region string
	size   string
}

func (v *CostView) Update(msg any) (string, []any) {
	m, ok := msg.(wizard.CollectedChangedMsg)
	if !ok {
		return v.last, nil
	}

	// Track collected values.
	if r, ok := m.Collected["region"].(string); ok {
		v.region = r
	}
	if s, ok := m.Collected["size"].(string); ok {
		v.size = s
	}

	// Can't calculate until we have both.
	if v.region == "" || v.size == "" {
		return v.last, nil
	}

	// Look up price.
	price, ok := prices[v.region][v.size]
	if !ok {
		return v.last, nil
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	v.last = style.Render(fmt.Sprintf("  Estimated cost: $%.2f/hr", price)) + "\n"

	// Publish CostUpdatedMsg — any view subscribed to this type will receive it.
	// We don't know (or care) who subscribes. The bus routes by type.
	return v.last, []any{CostUpdatedMsg{
		Region: v.region,
		Size:   v.size,
		Price:  price,
	}}
}

// Subscribe returns nil — CostView receives all engine broadcasts
// (StepChangedMsg, CollectedChangedMsg, StoreChangedMsg).
func (v *CostView) Subscribe() []reflect.Type {
	return nil
}

// --- QuotaView: reacts to CostView's CostUpdatedMsg ---

// QuotaView shows remaining instance quota for the selected region.
// It does NOT listen to engine broadcasts — it ONLY listens to CostUpdatedMsg
// published by CostView.
//
// This demonstrates "route by type": QuotaView doesn't know CostView exists.
// It just declares "I want CostUpdatedMsg" and the bus delivers it.
type QuotaView struct {
	last string
}

func (v *QuotaView) Update(msg any) (string, []any) {
	m, ok := msg.(CostUpdatedMsg)
	if !ok {
		return v.last, nil
	}

	remaining := quotas[m.Region]
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	v.last = style.Render(fmt.Sprintf("  Quota: %d instances remaining in %s", remaining, m.Region)) + "\n"
	return v.last, nil
}

// Subscribe declares that QuotaView wants CostUpdatedMsg.
// The bus sees this and routes any CostUpdatedMsg to this view.
// Without this, QuotaView would never receive CostUpdatedMsg.
func (v *QuotaView) Subscribe() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[CostUpdatedMsg](), // "route by type" — match on Go struct type
	}
}

// --- Main ---

func main() {
	ctx := context.Background()
	p := tui.Default()

	var region, size, name, replicas string

	flow := &wizard.Flow{
		Name: "deploy",

		// Layout defines the visual stack. Views render top-to-bottom,
		// above the active prompt.
		Layout: []wizard.ViewDef{
			// Built-in progress bar.
			{ID: "progress", View: wizard.NewProgressView(
				wizard.WithProgressGradient("#5A56E0", "#EE6FF8"),
			)},
			// Custom: shows cost after region+size are selected.
			{ID: "cost", View: &CostView{}},
			// Custom: shows quota, triggered by CostView's published message.
			{ID: "quota", View: &QuotaView{}},
		},

		Steps: []wizard.Step{
			{
				Name:        "region",
				Description: "Select region",
				Prompt:      wizard.SelectPrompt,
				Required:    true,
				Loader: wizard.StaticChoices(
					wizard.Choice{Label: "US East", Value: "us-east"},
					wizard.Choice{Label: "EU West", Value: "eu-west"},
				),
				Setter: func(v any) { region = v.(string) },
			},
			{
				Name:        "size",
				Description: "Select instance size",
				Prompt:      wizard.SelectPrompt,
				Required:    true,
				Loader: wizard.StaticChoices(
					wizard.Choice{Label: "Small (2 vCPU, 4GB)", Value: "small"},
					wizard.Choice{Label: "Medium (4 vCPU, 8GB)", Value: "medium"},
					wizard.Choice{Label: "Large (8 vCPU, 16GB)", Value: "large"},
				),
				Setter: func(v any) { size = v.(string) },
			},
			{
				Name:        "name",
				Description: "Instance name",
				Prompt:      wizard.TextInputPrompt,
				Required:    true,
				Setter:      func(v any) { name = v.(string) },
			},
			{
				Name:        "replicas",
				Description: "Number of replicas",
				Prompt:      wizard.TextInputPrompt,
				Required:    true,
				Default:     func(_ map[string]any) any { return "1" },
				Setter:      func(v any) { replicas = v.(string) },
			},
		},
	}

	fmt.Println("=== Deploy Instance (Views Demo) ===")
	fmt.Println()

	engine := wizard.NewEngine(p, nil)
	if err := engine.Run(ctx, flow); err != nil {
		log.Fatalf("Wizard failed: %v", err)
	}

	fmt.Println()
	fmt.Println("=== Result ===")
	fmt.Printf("  Region:   %s\n", region)
	fmt.Printf("  Size:     %s\n", size)
	fmt.Printf("  Name:     %s\n", name)
	fmt.Printf("  Replicas: %s\n", replicas)
}
