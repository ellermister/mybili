package main

import (
	"context"
	"log"
	"os"

	"mybili/internal/command"
	"mybili/internal/config"
)

func main() {
	cfg := config.Load()
	if err := command.Run(context.Background(), cfg, os.Args); err != nil {
		log.Fatal(err)
	}
}
