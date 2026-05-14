package exporter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportGhostty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.conf")

	if err := ExportGhostty(testTheme(), path); err != nil {
		t.Fatalf("ExportGhostty: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	for _, want := range []string{
		"palette = 0=#45475a",
		"palette = 15=#a6adc8",
		"background = #1e1e2e",
		"foreground = #cdd6f4",
		"cursor-color = #f5e0dc",
		"cursor-text = #1e1e2e",
		"selection-background = #585b70",
		"selection-foreground = #cdd6f4",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("missing %q in output:\n%s", want, content)
		}
	}
}

func TestGhosttyEscapeSequences(t *testing.T) {
	var buf bytes.Buffer
	WriteGhosttyEscapes(&buf, testTheme())
	out := buf.String()

	if len(out) == 0 {
		t.Fatal("empty escape sequence output")
	}
	if !strings.Contains(out, "\033]") {
		t.Error("missing OSC escape sequences")
	}
	// palette index 0
	if !strings.Contains(out, "4;0;rgb:45/47/5a") {
		t.Error("missing OSC 4 palette entry for index 0")
	}
	// background (OSC 11) + foreground (OSC 10) + cursor (OSC 12)
	if !strings.Contains(out, "10;rgb:cd/d6/f4") {
		t.Error("missing OSC 10 foreground")
	}
	if !strings.Contains(out, "11;rgb:1e/1e/2e") {
		t.Error("missing OSC 11 background")
	}
	if !strings.Contains(out, "12;rgb:f5/e0/dc") {
		t.Error("missing OSC 12 cursor")
	}
}

func TestGhosttyDefaultPath(t *testing.T) {
	path := GhosttyDefaultPath()
	if !strings.HasSuffix(path, filepath.Join(".config", "ghostty", "theme.conf")) {
		t.Errorf("unexpected path: %s", path)
	}
}
