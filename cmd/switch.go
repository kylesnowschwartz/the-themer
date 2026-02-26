package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kylesnowschwartz/the-themer/theme"
)

var switchThemesDir string

var switchCmd = &cobra.Command{
	Use:   "switch <theme-name>",
	Short: "Switch the active theme across all configured apps",
	Long: `Switch activates a theme by updating each app's active config.
This includes writing config pointers (theme.local, bat-theme.txt),
swapping symlinks (starship, fzf, eza), and invoking external tools
(nvim Themery, claude.json edit).

Only apps configured for the theme are switched. Others are skipped.

You can pass "dark" or "light" as the theme name to switch to the
default theme for that variant. Set defaults with:
  the-themer set dark <theme-name>
  the-themer set light <theme-name>`,
	Args: cobra.ExactArgs(1),
	RunE: runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
	switchCmd.Flags().StringVar(&switchThemesDir, "themes-dir", "./themes/", "path to the themes directory")
}

func runSwitch(cmd *cobra.Command, args []string) error {
	themeName := args[0]

	// Resolve "dark" and "light" to their configured default themes.
	if themeName == "dark" || themeName == "light" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolving home directory: %w", err)
		}
		resolved, err := theme.ReadDefault(home, themeName)
		if err != nil {
			return err
		}
		if resolved == "" {
			return fmt.Errorf("no default theme configured for %q — use \"the-themer set %s <theme-name>\" first", themeName, themeName)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Resolving %q to %q\n", themeName, resolved)
		themeName = resolved
	}

	t, err := theme.LoadTheme(switchThemesDir, themeName)
	if err != nil {
		return err
	}

	results := theme.Switch(t, theme.SwitchOpts{})

	var hasErrors bool
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  %s: ERROR %v\n", r.App, r.Err)
			hasErrors = true
		} else if r.Skipped {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: skipped (%s)\n", r.App, r.Message)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", r.App, r.Message)
		}
	}

	if hasErrors {
		return fmt.Errorf("some apps failed to switch")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "  state: WARNING could not resolve home: %v\n", err)
	} else if err := theme.WriteState(home, themeName); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "  state: WARNING could not write state: %v\n", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Switched to theme %q\n", themeName)
	return nil
}
