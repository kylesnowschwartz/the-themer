package hud_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kylesnowschwartz/the-themer/adapter"
	_ "github.com/kylesnowschwartz/the-themer/adapter/hud"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func TestGenerate_OracleBleu(t *testing.T) {
	cfg, err := palette.Load("../../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load bleu.toml: %v", err)
	}

	hud := adapter.ByName([]string{"hud"})
	if len(hud) != 1 {
		t.Fatalf("expected 1 hud adapter, got %d", len(hud))
	}

	got, err := hud[0].Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected, err := os.ReadFile("../../testdata/expected/hud/bleu.toml")
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
		if a.Name() == "hud" {
			found = true
			if a.DirName() != "hud" {
				t.Errorf("DirName: got %q, want %q", a.DirName(), "hud")
			}
			if a.FileName("bleu") != "bleu.toml" {
				t.Errorf("FileName: got %q, want %q", a.FileName("bleu"), "bleu.toml")
			}
		}
	}
	if !found {
		t.Fatal("hud adapter not registered")
	}
}
