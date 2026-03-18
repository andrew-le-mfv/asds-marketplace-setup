package main

import (
	"os"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/commands"
)

func main() {
	cmd := commands.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
