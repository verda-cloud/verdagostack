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
