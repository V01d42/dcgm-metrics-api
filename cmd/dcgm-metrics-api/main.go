package main

import (
	"log"

	"github.com/V01d42/dcgm-metrics-api/pkg/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
