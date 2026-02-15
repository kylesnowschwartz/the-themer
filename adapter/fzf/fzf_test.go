package fzf_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kylesnowschwartz/the-themer/adapter"
	_ "github.com/kylesnowschwartz/the-themer/adapter/fzf"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func TestGenerate_OracleBleu(t *testing.T) {
	cfg, err := palette.Load("../../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load bleu.toml: %v", err)
	}

	// fzf uses a per-adapter palette override; simulate what generate.go does.
	override, ok := cfg.Adapters["fzf"]
	if !ok {
		t.Fatal("expected fzf adapter override in bleu.toml")
	}
	adapterCfg := palette.Config{
		Theme:   cfg.Theme,
		Palette: override.Palette,
	}
	adapterCfg.ApplyDefaults()
	if err := adapterCfg.Validate(); err != nil {
		t.Fatalf("fzf override validation: %v", err)
	}

	fzf := adapter.ByName([]string{"fzf"})
	if len(fzf) != 1 {
		t.Fatalf("expected 1 fzf adapter, got %d", len(fzf))
	}

	got, err := fzf[0].Generate(adapterCfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected, err := os.ReadFile("../../testdata/expected/fzf/bleu.zsh")
	if err != nil {
		t.Fatalf("reading expected fixture: %v", err)
	}

	if !bytes.Equal(got, expected) {
		t.Errorf("output differs from oracle\n--- got ---\n%s\n--- want ---\n%s", got, expected)
	}
}

func TestAdapterRegistration(t *testing.T) {
	all := adapter.All()

	found := false
	for _, a := range all {
		if a.Name() == "fzf" {
			found = true
			if a.DirName() != "fzf" {
				t.Errorf("DirName: got %q, want %q", a.DirName(), "fzf")
			}
			if a.FileName("bleu") != "bleu.zsh" {
				t.Errorf("FileName: got %q, want %q", a.FileName("bleu"), "bleu.zsh")
			}
		}
	}
	if !found {
		t.Fatal("fzf adapter not registered")
	}
}
