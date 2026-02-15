package delta_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kylesnowschwartz/the-themer/adapter"
	_ "github.com/kylesnowschwartz/the-themer/adapter/delta"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func TestGenerate_OracleBleu(t *testing.T) {
	cfg, err := palette.Load("../../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load bleu.toml: %v", err)
	}

	// Delta uses the base palette directly -- no per-adapter override needed.
	delta := adapter.ByName([]string{"delta"})
	if len(delta) != 1 {
		t.Fatalf("expected 1 delta adapter, got %d", len(delta))
	}

	got, err := delta[0].Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected, err := os.ReadFile("../../testdata/expected/delta/bleu.gitconfig")
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
		if a.Name() == "delta" {
			found = true
			if a.DirName() != "delta" {
				t.Errorf("DirName: got %q, want %q", a.DirName(), "delta")
			}
			if a.FileName("bleu") != "bleu.gitconfig" {
				t.Errorf("FileName: got %q, want %q", a.FileName("bleu"), "bleu.gitconfig")
			}
		}
	}
	if !found {
		t.Fatal("delta adapter not registered")
	}
}
