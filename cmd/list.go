package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kylesnowschwartz/the-themer/theme"
)

var listThemesDir string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available themes with variant, active marker, and defaults",
	Long: `List scans the themes directory and prints every theme alongside its
variant, whether it is currently active, and whether it is the configured
dark or light default (see "the-themer set").`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listThemesDir, "themes-dir", defaultThemesDir(), "path to the themes directory")
}

func runList(cmd *cobra.Command, args []string) error {
	names, err := theme.ListThemes(listThemesDir)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home directory: %w", err)
	}

	current, err := theme.ReadState(home)
	if err != nil {
		return err
	}
	darkDefault, err := theme.ReadDefault(home, "dark")
	if err != nil {
		return err
	}
	lightDefault, err := theme.ReadDefault(home, "light")
	if err != nil {
		return err
	}

	for _, name := range names {
		variant := "?"
		if t, err := theme.LoadTheme(listThemesDir, name); err == nil {
			variant = t.Config.Theme.Variant
		}

		var markers []string
		if name == current {
			markers = append(markers, "current")
		}
		if name == darkDefault {
			markers = append(markers, "dark default")
		}
		if name == lightDefault {
			markers = append(markers, "light default")
		}

		line := fmt.Sprintf("%-22s %-5s", name, variant)
		if len(markers) > 0 {
			line += "  (" + strings.Join(markers, ", ") + ")"
		}
		fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(line, " "))
	}
	return nil
}
