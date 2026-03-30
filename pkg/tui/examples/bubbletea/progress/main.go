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

	// Default gradient progress bar
	prog, err := s.Progress(ctx, "Downloading artifacts")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i <= 20; i++ {
		prog.SetPercent(float64(i) / 20.0)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println()
	time.Sleep(300 * time.Millisecond)

	// Custom gradient colors
	prog, err = s.Progress(ctx, "Building image",
		tui.WithProgressGradient("#FF7CCB", "#FDFF8C"),
		tui.WithProgressWidth(50),
	)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i <= 30; i++ {
		prog.SetPercent(float64(i) / 30.0)
		time.Sleep(80 * time.Millisecond)
	}

	fmt.Println()
	time.Sleep(300 * time.Millisecond)

	// Solid fill, no percentage text
	prog, err = s.Progress(ctx, "Uploading",
		tui.WithProgressSolidFill("#04B575"),
		tui.WithoutPercent(),
	)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i <= 15; i++ {
		prog.Increment(1.0 / 15.0)
		time.Sleep(120 * time.Millisecond)
	}
}
