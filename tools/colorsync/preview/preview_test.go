package preview

import (
	"bytes"
	"testing"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func TestRender(t *testing.T) {
	theme := &palette.Theme{
		Name: "test", Background: "#1e1e2e", Foreground: "#cdd6f4", Cursor: "#f5e0dc",
		Colors: [16]string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	}

	var buf bytes.Buffer
	Render(&buf, theme)
	out := buf.String()

	if len(out) == 0 {
		t.Error("Render produced empty output")
	}
	if !bytes.Contains([]byte(out), []byte("test")) {
		t.Error("output should contain theme name")
	}
	if !bytes.Contains([]byte(out), []byte("\033[")) {
		t.Error("output should contain ANSI escape sequences")
	}
}
