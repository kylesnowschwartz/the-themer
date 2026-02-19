package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kylesnowschwartz/the-themer/theme"
)

var installThemesDir string

var installCmd = &cobra.Command{
	Use:   "install <theme-name>",
	Short: "Install a theme's configs to their filesystem destinations",
	Long: `Install copies per-app config files from the theme warehouse to the
locations where each application expects them.

Only apps with config directories in the theme are installed. Apps without
config in the theme are skipped.`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&installThemesDir, "themes-dir", "./themes/", "path to the themes directory")
}

func runInstall(cmd *cobra.Command, args []string) error {
	themeName := args[0]

	t, err := theme.LoadTheme(installThemesDir, themeName)
	if err != nil {
		return err
	}

	results := theme.Install(t, theme.InstallOpts{})

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
		return fmt.Errorf("some apps failed to install")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Installed theme %q\n", themeName)
	return nil
}
