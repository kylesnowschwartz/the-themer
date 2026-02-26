// Package cmd implements the CLI commands for the-themer.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "the-themer",
	Short: "Terminal theme warehouse — generate, install, and switch themes",
	Long: `the-themer manages terminal themes across multiple apps (ghostty,
bat, delta, fzf, starship, eza, gh-dash, neovim, claude).

Commands:
  generate   Render per-app configs from a palette TOML
  install    Deploy a theme's configs to the filesystem
  switch     Activate a theme across all configured apps
  set        Configure default themes for "dark" and "light" aliases

Set your defaults once, then switch by variant:
  the-themer set dark cobalt-next-neon
  the-themer set light dayfox
  the-themer switch dark`,
	SilenceUsage:  true, // don't dump usage on every RunE error
	SilenceErrors: true, // main.go handles error printing
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
