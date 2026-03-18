package main

import (
	"os"

	"github.com/your-org/asds-marketplace-setup/internal/commands"
)

func main() {
	cmd := commands.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
