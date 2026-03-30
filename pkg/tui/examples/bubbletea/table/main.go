package main

import (
	"context"
	"log"

	"github.com/verda-cloud/verdagostack/pkg/tui"
	_ "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea"
)

func main() {
	ctx := context.Background()
	s := tui.DefaultStatus()

	// Service status table
	err := s.Table(ctx,
		[]string{"Service", "Status", "Endpoint", "Uptime"},
		[][]string{
			{"apigateway", "healthy", "http://localhost:9090", "3d 14h"},
			{"apiserver", "healthy", "http://localhost:9091", "3d 14h"},
			{"postgres", "healthy", "localhost:5432", "5d 2h"},
			{"redis", "degraded", "localhost:6379", "1d 8h"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
