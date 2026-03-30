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

	// Password input (masked characters)
	secret, err := p.Password(ctx, "API token:")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Token length: %d\n", len(secret))
}
