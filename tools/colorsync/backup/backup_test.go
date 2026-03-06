package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupAndRestore_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "theme.conf")
	os.WriteFile(original, []byte("original content"), 0644)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	BeginApply()
	if err := SaveBackup(original); err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	os.WriteFile(original, []byte("new content"), 0644)

	actions, err := Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(actions) == 0 {
		t.Error("expected restore actions")
	}

	data, _ := os.ReadFile(original)
	if string(data) != "original content" {
		t.Errorf("got %q, want %q", string(data), "original content")
	}
}

func TestBackupAndRestore_NewFile(t *testing.T) {
	dir := t.TempDir()
	newFile := filepath.Join(dir, "newfile.lua")

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	BeginApply()
	if err := SaveBackup(newFile); err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	os.WriteFile(newFile, []byte("new theme"), 0644)

	actions, err := Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(actions) == 0 {
		t.Error("expected restore actions")
	}

	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Error("expected file to be deleted after undo")
	}
}

func TestRestoreNothingToUndo(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	_, err := Restore()
	if err == nil {
		t.Error("expected error when nothing to undo")
	}
}

func TestStackMultipleUndos(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "theme.conf")
	os.WriteFile(file, []byte("v1"), 0644)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	// Apply 1: v1 -> v2
	BeginApply()
	SaveBackup(file)
	os.WriteFile(file, []byte("v2"), 0644)

	// Apply 2: v2 -> v3
	BeginApply()
	SaveBackup(file)
	os.WriteFile(file, []byte("v3"), 0644)

	// Apply 3: v3 -> v4
	BeginApply()
	SaveBackup(file)
	os.WriteFile(file, []byte("v4"), 0644)

	if Depth() != 3 {
		t.Fatalf("expected depth 3, got %d", Depth())
	}

	// Undo 1: v4 -> v3
	Restore()
	data, _ := os.ReadFile(file)
	if string(data) != "v3" {
		t.Errorf("after undo 1: got %q, want %q", string(data), "v3")
	}
	if Depth() != 2 {
		t.Errorf("expected depth 2, got %d", Depth())
	}

	// Undo 2: v3 -> v2
	Restore()
	data, _ = os.ReadFile(file)
	if string(data) != "v2" {
		t.Errorf("after undo 2: got %q, want %q", string(data), "v2")
	}

	// Undo 3: v2 -> v1
	Restore()
	data, _ = os.ReadFile(file)
	if string(data) != "v1" {
		t.Errorf("after undo 3: got %q, want %q", string(data), "v1")
	}

	// No more undos
	if Depth() != 0 {
		t.Errorf("expected depth 0, got %d", Depth())
	}
	_, err := Restore()
	if err == nil {
		t.Error("expected error when stack empty")
	}
}
