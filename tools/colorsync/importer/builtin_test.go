package importer

import (
	"testing"
)

func TestGetBuiltin(t *testing.T) {
	theme, err := GetBuiltin("catppuccin-mocha")
	if err != nil {
		t.Fatalf("GetBuiltin: %v", err)
	}
	if theme.Name != "catppuccin-mocha" {
		t.Errorf("Name: got %q", theme.Name)
	}
	if theme.Background != "#1e1e2e" {
		t.Errorf("Background: got %q", theme.Background)
	}
	for i, c := range theme.Colors {
		if c == "" {
			t.Errorf("color%d is empty", i)
		}
	}
}

func TestGetBuiltinUnknown(t *testing.T) {
	_, err := GetBuiltin("nonexistent")
	if err == nil {
		t.Error("expected error for unknown theme")
	}
}

func TestListBuiltins(t *testing.T) {
	names := ListBuiltins()
	if len(names) < 6 {
		t.Errorf("expected at least 6 builtins, got %d", len(names))
	}
}
