package ghostty_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kylesnowschwartz/the-themer/adapter"
	_ "github.com/kylesnowschwartz/the-themer/adapter/ghostty"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func TestGenerate_OracleBleu(t *testing.T) {
	cfg, err := palette.Load("../../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load bleu.toml: %v", err)
	}

	ghostty := adapter.ByName([]string{"ghostty"})
	if len(ghostty) != 1 {
		t.Fatalf("expected 1 ghostty adapter, got %d", len(ghostty))
	}

	got, err := ghostty[0].Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected, err := os.ReadFile("../../testdata/expected/ghostty/bleu")
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
		if a.Name() == "ghostty" {
			found = true
			if a.DirName() != "ghostty" {
				t.Errorf("DirName: got %q, want %q", a.DirName(), "ghostty")
			}
			if a.FileName("bleu") != "bleu" {
				t.Errorf("FileName: got %q, want %q", a.FileName("bleu"), "bleu")
			}
		}
	}
	if !found {
		t.Fatal("ghostty adapter not registered")
	}
}
