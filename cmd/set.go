package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kylesnowschwartz/the-themer/theme"
)

var setThemesDir string

var setCmd = &cobra.Command{
	Use:   "set <dark|light> <theme-name>",
	Short: "Set the default theme for a variant (dark or light)",
	Long: `Set configures which theme to use when switching by variant name.

After setting defaults, "the-themer switch dark" and "the-themer switch light"
resolve to the configured theme names.

Examples:
  the-themer set dark cobalt-next-neon
  the-themer set light dayfox`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().StringVar(&setThemesDir, "themes-dir", "./themes/", "path to the themes directory")
}

func runSet(cmd *cobra.Command, args []string) error {
	variant := args[0]
	themeName := args[1]

	if variant != "dark" && variant != "light" {
		return fmt.Errorf("variant must be %q or %q, got %q", "dark", "light", variant)
	}

	// Validate the theme exists.
	if _, err := theme.LoadTheme(setThemesDir, themeName); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home directory: %w", err)
	}

	if err := theme.WriteDefault(home, variant, themeName); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set %s default to %q\n", variant, themeName)
	return nil
}
