// Package cmd implements the CLI commands for the-themer.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "the-themer",
	Short: "Generate themed config files from a TOML color palette",
	Long: `the-themer takes a TOML color palette and generates per-app config
files for terminal applications like ghostty, starship, and neovim.

One palette in, themed configs out.`,
	SilenceUsage:  true, // don't dump usage on every RunE error
	SilenceErrors: true, // main.go handles error printing
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
