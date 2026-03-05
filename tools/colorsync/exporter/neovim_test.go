package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func testTheme() *palette.Theme {
	return &palette.Theme{
		Name: "test-theme", Background: "#1e1e2e", Foreground: "#cdd6f4", Cursor: "#f5e0dc",
		Colors: [16]string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	}
}

func TestExportNeovim(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-theme.lua")

	err := ExportNeovim(testTheme(), path)
	if err != nil {
		t.Fatalf("ExportNeovim: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "vim.o.background") {
		t.Error("missing vim.o.background")
	}
	if !strings.Contains(content, "hi clear") {
		t.Error("missing hi clear")
	}
	if !strings.Contains(content, "Normal") {
		t.Error("missing Normal highlight group")
	}
	if !strings.Contains(content, "#1e1e2e") {
		t.Error("missing background color")
	}
	if !strings.Contains(content, "#cdd6f4") {
		t.Error("missing foreground color")
	}
}

func TestNeovimDefaultPath(t *testing.T) {
	path := NeovimDefaultPath("test-theme")
	if !strings.Contains(path, filepath.Join(".config", "nvim", "colors", "test-theme.lua")) {
		t.Errorf("unexpected path: %s", path)
	}
}

func TestFormatNeovimActivation(t *testing.T) {
	got := FormatNeovimActivation("test-theme")
	want := "vim.cmd('colorscheme test_theme')"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
