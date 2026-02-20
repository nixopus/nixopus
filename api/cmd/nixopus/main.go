package main

import (
	"os"

	"github.com/raghavyuva/nixopus-api/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
