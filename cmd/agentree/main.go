// Package main is the entry point for the agentree CLI application.
package main

import (
	"os"

	"github.com/AryaLabsHQ/agentree/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
