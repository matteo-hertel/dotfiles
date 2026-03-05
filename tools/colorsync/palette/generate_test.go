package palette

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	theme, err := Generate("test-gen", "#1a1b26", "#c0caf5", "#7aa2f7")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if theme.Name != "test-gen" {
		t.Errorf("Name: got %q", theme.Name)
	}
	if theme.Background != "#1a1b26" {
		t.Errorf("Background: got %q", theme.Background)
	}
	if theme.Foreground != "#c0caf5" {
		t.Errorf("Foreground: got %q", theme.Foreground)
	}

	// All 16 colors should be valid hex
	for i, c := range theme.Colors {
		if len(c) != 7 || c[0] != '#' {
			t.Errorf("color%d invalid hex: %q", i, c)
		}
	}

	// color1 (red) should have high red component
	r, _, _, _ := ParseHex(theme.Colors[1])
	if r < 150 {
		t.Errorf("color1 (red) should have high red, got R=%d", r)
	}

	// Bright variants (8-15) should be brighter than normal (0-7)
	for i := 0; i < 8; i++ {
		_, _, _, err1 := ParseHex(theme.Colors[i])
		_, _, _, err2 := ParseHex(theme.Colors[i+8])
		if err1 != nil || err2 != nil {
			t.Errorf("invalid color at %d or %d", i, i+8)
		}
	}
}

func TestGenerateInvalidHex(t *testing.T) {
	_, err := Generate("bad", "invalid", "#ffffff", "#ffffff")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}
