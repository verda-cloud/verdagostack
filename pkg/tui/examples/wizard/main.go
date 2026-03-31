// Command wizard demonstrates the TUI wizard engine with a multi-step
// "deploy service" flow. It uses mock data to simulate API responses,
// conditional steps, dependencies, and back-navigation.
//
// Run: go run ./pkg/tui/examples/wizard
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	_ "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea"
	"github.com/verda-cloud/verdagostack/pkg/tui/wizard"
)

// Mock data simulating API responses.
var (
	environments = []wizard.Choice{
		{Label: "Production", Value: "prod", Description: "Live traffic"},
		{Label: "Staging", Value: "staging", Description: "Pre-release testing"},
		{Label: "Development", Value: "dev", Description: "Local development"},
	}

	// Regions available per environment (dependency example).
	regionsByEnv = map[string][]wizard.Choice{
		"prod": {
			{Label: "US East (us-east-1)", Value: "us-east-1"},
			{Label: "EU West (eu-west-1)", Value: "eu-west-1"},
			{Label: "AP Southeast (ap-southeast-1)", Value: "ap-southeast-1"},
		},
		"staging": {
			{Label: "US East (us-east-1)", Value: "us-east-1"},
			{Label: "EU West (eu-west-1)", Value: "eu-west-1"},
		},
		"dev": {}, // empty — will trigger auto-back
	}

	// Instance sizes available per region.
	sizesByRegion = map[string][]wizard.Choice{
		"us-east-1": {
			{Label: "Small (2 vCPU, 4GB)", Value: "small"},
			{Label: "Medium (4 vCPU, 8GB)", Value: "medium"},
			{Label: "Large (8 vCPU, 16GB)", Value: "large"},
			{Label: "XLarge (16 vCPU, 32GB)", Value: "xlarge"},
		},
		"eu-west-1": {
			{Label: "Small (2 vCPU, 4GB)", Value: "small"},
			{Label: "Medium (4 vCPU, 8GB)", Value: "medium"},
		},
		"ap-southeast-1": {
			{Label: "Small (2 vCPU, 4GB)", Value: "small"},
		},
	}

	runtimes = []wizard.Choice{
		{Label: "Go 1.25", Value: "go-1.25"},
		{Label: "Node.js 22 LTS", Value: "node-22"},
		{Label: "Python 3.13", Value: "python-3.13"},
		{Label: "Rust 1.85", Value: "rust-1.85"},
		{Label: "Java 21 LTS", Value: "java-21"},
	}

	databases = []wizard.Choice{
		{Label: "None", Value: ""},
		{Label: "PostgreSQL 17", Value: "postgres-17"},
		{Label: "MySQL 9", Value: "mysql-9"},
		{Label: "CockroachDB", Value: "cockroachdb"},
		{Label: "Valkey 8", Value: "valkey-8"},
	}
)

func main() {
	ctx := context.Background()
	p := tui.Default()

	// Collected values — in a real CLI these would be fields on an options struct.
	var env, region, size, runtime, db, serviceName, replicas string
	var enableTLS bool

	flow := &wizard.Flow{
		Name: "deploy-service",
		Steps: []wizard.Step{
			{
				Name:        "environment",
				Description: "Select target environment",
				Prompt:      wizard.SelectPrompt,
				Required:    true,
				Loader:      wizard.StaticChoices(environments...),
				Setter:      func(v any) { env = v.(string) },
			},
			{
				Name:        "region",
				Description: "Select deployment region",
				Prompt:      wizard.SelectPrompt,
				Required:    true,
				DependsOn:   []string{"environment"},
				Loader: func(_ context.Context, _ tui.Prompter, c map[string]any) ([]wizard.Choice, error) {
					// Simulates an API call filtered by environment.
					return regionsByEnv[c["environment"].(string)], nil
				},
				Setter: func(v any) { region = v.(string) },
			},
			{
				Name:        "size",
				Description: "Select instance size",
				Prompt:      wizard.SelectPrompt,
				Required:    true,
				DependsOn:   []string{"region"},
				Loader: func(_ context.Context, _ tui.Prompter, c map[string]any) ([]wizard.Choice, error) {
					// Simulates API: available sizes depend on region.
					return sizesByRegion[c["region"].(string)], nil
				},
				Setter: func(v any) { size = v.(string) },
			},
			{
				Name:        "runtime",
				Description: "Select application runtime",
				Prompt:      wizard.SelectPrompt,
				Required:    true,
				Loader:      wizard.StaticChoices(runtimes...),
				Setter:      func(v any) { runtime = v.(string) },
			},
			{
				Name:        "database",
				Description: "Select database (optional)",
				Prompt:      wizard.SelectPrompt,
				Required:    false,
				Loader:      wizard.StaticChoices(databases...),
				Setter:      func(v any) { db = v.(string) },
			},
			{
				Name:        "enable-tls",
				Description: "Enable TLS?",
				Prompt:      wizard.ConfirmPrompt,
				Required:    true,
				// Skip TLS prompt for dev — always false.
				ShouldSkip: func(c map[string]any) bool { return c["environment"] == "dev" },
				Setter:     func(v any) { enableTLS = v.(bool) },
			},
			{
				Name:        "replicas",
				Description: "Number of replicas",
				Prompt:      wizard.TextInputPrompt,
				Required:    true,
				Default: func(c map[string]any) any {
					// Smart default based on environment.
					switch c["environment"] {
					case "prod":
						return "3"
					case "staging":
						return "2"
					default:
						return "1"
					}
				},
				Setter: func(v any) { replicas = v.(string) },
			},
			{
				Name:        "service-name",
				Description: "Service name",
				Prompt:      wizard.TextInputPrompt,
				Required:    true,
				Default: func(c map[string]any) any {
					// Auto-generate a name from runtime + random suffix.
					rt := c["runtime"].(string)
					prefix := strings.Split(rt, "-")[0]
					return fmt.Sprintf("%s-svc-%04d", prefix, rand.IntN(10000)) //nolint:gosec // example only
				},
				Setter: func(v any) { serviceName = v.(string) },
			},
		},
	}

	fmt.Println("=== Deploy Service Wizard ===")
	fmt.Println("Navigate: ↑/↓ to move, Enter to select, Esc to go back")
	fmt.Println()

	engine := wizard.NewEngine(p)
	if err := engine.Run(ctx, flow); err != nil {
		log.Fatalf("Wizard failed: %v", err)
	}

	// Summary
	fmt.Println()
	fmt.Println("=== Deployment Summary ===")
	fmt.Printf("  Service:     %s\n", serviceName)
	fmt.Printf("  Environment: %s\n", env)
	fmt.Printf("  Region:      %s\n", region)
	fmt.Printf("  Size:        %s\n", size)
	fmt.Printf("  Runtime:     %s\n", runtime)
	if db != "" {
		fmt.Printf("  Database:    %s\n", db)
	}
	fmt.Printf("  TLS:         %v\n", enableTLS)
	fmt.Printf("  Replicas:    %s\n", replicas)
}
