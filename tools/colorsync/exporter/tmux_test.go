package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportTmux(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.conf")

	err := ExportTmux(testTheme(), path)
	if err != nil {
		t.Fatalf("ExportTmux: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "status-style") {
		t.Error("missing status-style")
	}
	if !strings.Contains(content, "pane-border-style") {
		t.Error("missing pane-border-style")
	}
	if !strings.Contains(content, "#1e1e2e") {
		t.Error("missing background color")
	}
}

func TestTmuxDefaultPath(t *testing.T) {
	path := TmuxDefaultPath()
	if !strings.Contains(path, filepath.Join(".tmux", "theme.conf")) {
		t.Errorf("unexpected path: %s", path)
	}
}
