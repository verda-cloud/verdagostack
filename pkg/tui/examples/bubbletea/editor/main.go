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

	// Editor with default content and file extension hint
	content, err := p.Editor(ctx, "Edit configuration:",
		tui.WithEditorDefault("# Service Configuration\nname: my-service\nport: 8080\n"),
		tui.WithFileExt(".yaml"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Configuration:\n%s\n", content)
}
