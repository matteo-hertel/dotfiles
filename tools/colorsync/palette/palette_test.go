package palette

import (
	"path/filepath"
	"testing"
)

func TestThemeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	theme := Theme{
		Name:       "test-theme",
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Cursor:     "#f5e0dc",
		Colors: [16]string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	}

	path := filepath.Join(dir, "test-theme.json")
	if err := theme.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != theme.Name {
		t.Errorf("Name: got %q, want %q", loaded.Name, theme.Name)
	}
	if loaded.Background != theme.Background {
		t.Errorf("Background: got %q, want %q", loaded.Background, theme.Background)
	}
	if loaded.Colors != theme.Colors {
		t.Errorf("Colors mismatch")
	}
}

func TestLoadThemesDir(t *testing.T) {
	dir := t.TempDir()
	theme := Theme{Name: "alpha", Background: "#000000", Foreground: "#ffffff", Cursor: "#ffffff"}
	theme.Save(filepath.Join(dir, "alpha.json"))

	theme2 := Theme{Name: "beta", Background: "#111111", Foreground: "#eeeeee", Cursor: "#eeeeee"}
	theme2.Save(filepath.Join(dir, "beta.json"))

	themes, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("got %d themes, want 2", len(themes))
	}
}

func TestParseHexColor(t *testing.T) {
	r, g, b, err := ParseHex("#ff8800")
	if err != nil {
		t.Fatalf("ParseHex: %v", err)
	}
	if r != 255 || g != 136 || b != 0 {
		t.Errorf("got (%d,%d,%d), want (255,136,0)", r, g, b)
	}

	_, _, _, err = ParseHex("invalid")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}
