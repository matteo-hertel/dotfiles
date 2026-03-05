package exporter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportItermFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.itermcolors")

	err := ExportItermFile(testTheme(), path)
	if err != nil {
		t.Fatalf("ExportItermFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Background Color") {
		t.Error("missing Background Color")
	}
	if !strings.Contains(content, "Foreground Color") {
		t.Error("missing Foreground Color")
	}
	if !strings.Contains(content, "Ansi 0 Color") {
		t.Error("missing Ansi 0 Color")
	}
	if !strings.Contains(content, "plist") {
		t.Error("missing plist header")
	}
}

func TestItermEscapeSequences(t *testing.T) {
	var buf bytes.Buffer
	WriteItermEscapes(&buf, testTheme())
	out := buf.String()

	if len(out) == 0 {
		t.Error("empty escape sequence output")
	}
	if !strings.Contains(out, "\033]") {
		t.Error("missing OSC escape sequences")
	}
}
