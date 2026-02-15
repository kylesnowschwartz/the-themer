package main

import (
	"fmt"
	"os"

	"github.com/kylesnowschwartz/the-themer/cmd"
	// Import adapter packages here to register them via init().
	// Phase 2+ will add lines like:
	// _ "github.com/kylesnowschwartz/the-themer/adapter/ghostty"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
