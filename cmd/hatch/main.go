// Package main is the entry point for the hatch CLI application.
package main

import (
	"os"
	
	"github.com/AryaLabsHQ/hatch/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}