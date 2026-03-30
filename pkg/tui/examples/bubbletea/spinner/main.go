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
