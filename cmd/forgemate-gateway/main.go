package main

import (
	"context"
	"log"

	"forgemate/internal/app"
	"forgemate/internal/config"
)

func main() {
	cfg := config.Load()
	if err := app.RunGateway(context.Background(), cfg); err != nil {
		log.Fatalf("gateway failed: %v", err)
	}
}
