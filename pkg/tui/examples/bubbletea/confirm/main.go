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
