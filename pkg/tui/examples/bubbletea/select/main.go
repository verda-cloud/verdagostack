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
