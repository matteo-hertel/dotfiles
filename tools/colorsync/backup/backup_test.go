package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupAndRestore_ExistingFile(t *testing.T) {
	// Set up: create a temp "original" file
	dir := t.TempDir()
	original := filepath.Join(dir, "theme.conf")
	os.WriteFile(original, []byte("original content"), 0644)

	// Override backup dir for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	// Backup
	err := SaveBackup(original)
	if err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	// Overwrite original (simulating apply)
	os.WriteFile(original, []byte("new content"), 0644)

	// Restore
	actions, err := Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(actions) == 0 {
		t.Error("expected restore actions")
	}

	// Verify original content is back
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

	// Backup a file that doesn't exist yet
	err := SaveBackup(newFile)
	if err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	// Create the file (simulating apply)
	os.WriteFile(newFile, []byte("new theme"), 0644)

	// Restore should delete it
	actions, err := Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(actions) == 0 {
		t.Error("expected restore actions")
	}

	// File should not exist
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
