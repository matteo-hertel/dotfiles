package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportP10k(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshtheme")

	// Write a minimal existing config
	initial := "source /opt/homebrew/opt/powerlevel10k/share/powerlevel10k/powerlevel10k.zsh-theme\nPOWERLEVEL9K_PROMPT_ON_NEWLINE=true\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := ExportP10k(testTheme(), path)
	if err != nil {
		t.Fatalf("ExportP10k: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	// Should contain original content
	if !strings.Contains(content, "POWERLEVEL9K_PROMPT_ON_NEWLINE=true") {
		t.Error("missing original config content")
	}

	// Should contain markers
	if !strings.Contains(content, "# --- colorsync theme start ---") {
		t.Error("missing start marker")
	}
	if !strings.Contains(content, "# --- colorsync theme end ---") {
		t.Error("missing end marker")
	}

	// Should contain expected POWERLEVEL9K variables
	expectedVars := []string{
		"POWERLEVEL9K_CUSTOM_USER_BACKGROUND",
		"POWERLEVEL9K_CUSTOM_USER_FOREGROUND",
		"POWERLEVEL9K_DIR_BACKGROUND",
		"POWERLEVEL9K_DIR_FOREGROUND",
		"POWERLEVEL9K_VCS_CLEAN_BACKGROUND",
		"POWERLEVEL9K_VCS_MODIFIED_BACKGROUND",
		"POWERLEVEL9K_VCS_UNTRACKED_BACKGROUND",
		"POWERLEVEL9K_STATUS_OK_FOREGROUND",
		"POWERLEVEL9K_DATE_BACKGROUND",
		"POWERLEVEL9K_TIME_BACKGROUND",
		"POWERLEVEL9K_VI_MODE_NORMAL_BACKGROUND",
		"POWERLEVEL9K_VI_MODE_INSERT_BACKGROUND",
		"POWERLEVEL9K_VI_MODE_VISUAL_BACKGROUND",
	}
	for _, v := range expectedVars {
		if !strings.Contains(content, v) {
			t.Errorf("missing variable: %s", v)
		}
	}

	// Should contain theme colors
	if !strings.Contains(content, "#89b4fa") { // blue (Colors[4])
		t.Error("missing blue accent color")
	}
	if !strings.Contains(content, "#585b70") { // bright black (Colors[8])
		t.Error("missing bright black color")
	}
}

func TestExportP10kReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshtheme")

	initial := "# before\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Apply once
	if err := ExportP10k(testTheme(), path); err != nil {
		t.Fatalf("first ExportP10k: %v", err)
	}

	// Apply again — should replace, not duplicate
	if err := ExportP10k(testTheme(), path); err != nil {
		t.Fatalf("second ExportP10k: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	// Count markers — should appear exactly once each
	startCount := strings.Count(content, "# --- colorsync theme start ---")
	endCount := strings.Count(content, "# --- colorsync theme end ---")

	if startCount != 1 {
		t.Errorf("expected 1 start marker, got %d", startCount)
	}
	if endCount != 1 {
		t.Errorf("expected 1 end marker, got %d", endCount)
	}

	// Original content should still be there
	if !strings.Contains(content, "# before") {
		t.Error("missing original content after replacement")
	}
}

func TestP10kDefaultPath(t *testing.T) {
	path := P10kDefaultPath()
	if !strings.HasSuffix(path, ".zshtheme") {
		t.Errorf("unexpected path: %s", path)
	}
}

func TestGenerateP10kBlock(t *testing.T) {
	block := GenerateP10kBlock(testTheme())

	if !strings.HasPrefix(block, "# --- colorsync theme start ---") {
		t.Error("block does not start with start marker")
	}
	if !strings.HasSuffix(strings.TrimRight(block, "\n"), "# --- colorsync theme end ---") {
		t.Error("block does not end with end marker")
	}

	// Verify color mapping: custom user bg should be Colors[4] (blue)
	if !strings.Contains(block, "POWERLEVEL9K_CUSTOM_USER_BACKGROUND='#89b4fa'") {
		t.Error("custom user bg should be Colors[4] (blue)")
	}
	// Custom user fg should be background
	if !strings.Contains(block, "POWERLEVEL9K_CUSTOM_USER_FOREGROUND='#1e1e2e'") {
		t.Error("custom user fg should be theme background")
	}
	// Dir bg should be Colors[8] (bright black)
	if !strings.Contains(block, "POWERLEVEL9K_DIR_BACKGROUND='#585b70'") {
		t.Error("dir bg should be Colors[8] (bright black)")
	}
}
