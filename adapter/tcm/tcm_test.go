package tcm_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kylesnowschwartz/the-themer/adapter"
	_ "github.com/kylesnowschwartz/the-themer/adapter/tcm"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func TestGenerate_OracleBleu(t *testing.T) {
	cfg, err := palette.Load("../../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load bleu.toml: %v", err)
	}

	tcms := adapter.ByName([]string{"tcm"})
	if len(tcms) != 1 {
		t.Fatalf("expected 1 tcm adapter, got %d", len(tcms))
	}

	got, err := tcms[0].Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected, err := os.ReadFile("../../testdata/expected/tcm/bleu.json")
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
		if a.Name() == "tcm" {
			found = true
			if a.DirName() != "tcm" {
				t.Errorf("DirName: got %q, want %q", a.DirName(), "tcm")
			}
			if a.FileName("bleu") != "bleu.json" {
				t.Errorf("FileName: got %q, want %q", a.FileName("bleu"), "bleu.json")
			}
		}
	}
	if !found {
		t.Fatal("tcm adapter not registered")
	}
}

// TestGenerate_AdapterOverride ensures the [adapters.tcm.palette] override
// path goes through ApplyDefaults + Validate just like the input palette.
// We synthesize a Config with only tcm tokens overridden and verify the
// generated JSON reflects those overrides.
func TestGenerate_AdapterOverride(t *testing.T) {
	cfg, err := palette.Load("../../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load bleu.toml: %v", err)
	}

	// Synthesize an override that only changes a few tokens. The full
	// palette is required (Validate enforces all 16 ANSI + bg/fg).
	override := cfg.Palette
	override.UI.Accent = "#123456" // changes tcm "blue"
	override.UI.Warning = "#abcdef"

	adapterCfg := palette.Config{
		Theme:   cfg.Theme,
		Palette: override,
	}
	adapterCfg.ApplyDefaults()
	if err := adapterCfg.Validate(); err != nil {
		t.Fatalf("override validation: %v", err)
	}

	tcms := adapter.ByName([]string{"tcm"})
	got, err := tcms[0].Generate(adapterCfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if !bytes.Contains(got, []byte(`"blue": "#123456"`)) {
		t.Errorf("override accent did not propagate to blue token\n%s", got)
	}
	if !bytes.Contains(got, []byte(`"yellow": "#abcdef"`)) {
		t.Errorf("override warning did not propagate to yellow token\n%s", got)
	}
}
