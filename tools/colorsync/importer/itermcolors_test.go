package importer

import (
	"testing"
)

func TestParseItermColors(t *testing.T) {
	theme, err := ParseItermColors("../testdata/test.itermcolors")
	if err != nil {
		t.Fatalf("ParseItermColors: %v", err)
	}

	if theme.Name != "test" {
		t.Errorf("Name: got %q, want %q", theme.Name, "test")
	}
	if theme.Background != "#1e1e2e" {
		t.Errorf("Background: got %q, want %q", theme.Background, "#1e1e2e")
	}
	if theme.Foreground != "#cdd6f4" {
		t.Errorf("Foreground: got %q, want %q", theme.Foreground, "#cdd6f4")
	}
	if theme.Cursor != "#f5e0dc" {
		t.Errorf("Cursor: got %q, want %q", theme.Cursor, "#f5e0dc")
	}

	wantColors := [16]string{
		"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
		"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
	}
	for i, want := range wantColors {
		if theme.Colors[i] != want {
			t.Errorf("Color%d: got %q, want %q", i, theme.Colors[i], want)
		}
	}
}
