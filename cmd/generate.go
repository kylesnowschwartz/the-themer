package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kylesnowschwartz/the-themer/adapter"
	"github.com/kylesnowschwartz/the-themer/palette"
)

var (
	inputFlag    string
	outputFlag   string
	adaptersFlag string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate themed config files from a TOML palette",
	Long: `Parse a TOML color palette file, apply defaults, validate, and
generate themed config files for each registered adapter.`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&inputFlag, "input", "i", "", "path to TOML palette file (required)")
	generateCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output directory (default: ./{theme-name}-theme/)")
	generateCmd.Flags().StringVar(&adaptersFlag, "adapters", "", "comma-separated list of adapters to run (default: all)")

	generateCmd.MarkFlagRequired("input")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	cfg, err := palette.Load(inputFlag)
	if err != nil {
		return err
	}

	outDir := outputFlag
	if outDir == "" {
		outDir = fmt.Sprintf("./%s-theme", cfg.Theme.Name)
	}

	var selected []adapter.Adapter
	if adaptersFlag != "" {
		names := strings.Split(adaptersFlag, ",")
		selected = adapter.ByName(names)
	} else {
		selected = adapter.All()
	}

	if len(selected) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No adapters registered. Nothing to generate.")
		return nil
	}

	for _, a := range selected {
		content, err := a.Generate(cfg)
		if err != nil {
			return fmt.Errorf("adapter %s: %w", a.Name(), err)
		}

		dir := filepath.Join(outDir, a.DirName())
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		filePath := filepath.Join(dir, a.FileName(cfg.Theme.Name))
		if err := os.WriteFile(filePath, content, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", filePath, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  %s -> %s\n", a.Name(), filePath)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Generated %d file(s) in %s\n", len(selected), outDir)
	return nil
}
