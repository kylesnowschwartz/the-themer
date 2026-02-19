package bat_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kylesnowschwartz/the-themer/adapter"
	_ "github.com/kylesnowschwartz/the-themer/adapter/bat"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func TestGenerate_OracleBleu(t *testing.T) {
	cfg, err := palette.Load("../../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load bleu.toml: %v", err)
	}

	bat := adapter.ByName([]string{"bat"})
	if len(bat) != 1 {
		t.Fatalf("expected 1 bat adapter, got %d", len(bat))
	}

	got, err := bat[0].Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected, err := os.ReadFile("../../testdata/expected/bat/bleu.tmTheme")
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
		if a.Name() == "bat" {
			found = true
			if a.DirName() != "bat" {
				t.Errorf("DirName: got %q, want %q", a.DirName(), "bat")
			}
			if a.FileName("bleu") != "bleu.tmTheme" {
				t.Errorf("FileName: got %q, want %q", a.FileName("bleu"), "bleu.tmTheme")
			}
		}
	}
	if !found {
		t.Fatal("bat adapter not registered")
	}
}
