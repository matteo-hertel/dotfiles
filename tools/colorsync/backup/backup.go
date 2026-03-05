package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest tracks what was backed up so undo can restore it
type Manifest struct {
	// Map of original file path -> backup info
	Files map[string]FileBackup `json:"files"`
	// The theme name that was previously active (if any)
	PreviousTheme string `json:"previous_theme,omitempty"`
}

type FileBackup struct {
	// BackupPath is where the copy lives in the backup dir, empty if file didn't exist
	BackupPath string `json:"backup_path,omitempty"`
	// Existed indicates whether the file existed before apply
	Existed bool `json:"existed"`
}

func BackupDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "colorsync", "backup")
}

func manifestPath() string {
	return filepath.Join(BackupDir(), "manifest.json")
}

// SaveBackup copies the current file to the backup dir before it gets overwritten.
// Call this for each file BEFORE writing the new version.
func SaveBackup(originalPath string) error {
	dir := BackupDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	manifest := loadManifest()

	info, err := os.Stat(originalPath)
	if os.IsNotExist(err) {
		// File doesn't exist yet -- record that so undo can delete it
		manifest.Files[originalPath] = FileBackup{Existed: false}
		return saveManifest(manifest)
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("cannot backup directory: %s", originalPath)
	}

	// Copy file to backup dir with a safe name
	backupName := safeFileName(originalPath)
	backupPath := filepath.Join(dir, backupName)

	data, err := os.ReadFile(originalPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return err
	}

	manifest.Files[originalPath] = FileBackup{
		BackupPath: backupPath,
		Existed:    true,
	}
	return saveManifest(manifest)
}

// Restore puts everything back to the way it was before the last apply.
// Returns the list of actions taken.
func Restore() ([]string, error) {
	manifest := loadManifest()
	if len(manifest.Files) == 0 {
		return nil, fmt.Errorf("nothing to undo (no backup found)")
	}

	var actions []string

	for originalPath, info := range manifest.Files {
		if !info.Existed {
			// File was created by apply -- delete it
			if err := os.Remove(originalPath); err != nil && !os.IsNotExist(err) {
				return actions, fmt.Errorf("removing %s: %w", originalPath, err)
			}
			actions = append(actions, fmt.Sprintf("Removed %s (was newly created)", originalPath))
		} else {
			// File existed before -- restore the backup
			data, err := os.ReadFile(info.BackupPath)
			if err != nil {
				return actions, fmt.Errorf("reading backup %s: %w", info.BackupPath, err)
			}
			if err := os.MkdirAll(filepath.Dir(originalPath), 0755); err != nil {
				return actions, err
			}
			if err := os.WriteFile(originalPath, data, 0644); err != nil {
				return actions, fmt.Errorf("restoring %s: %w", originalPath, err)
			}
			actions = append(actions, fmt.Sprintf("Restored %s", originalPath))
		}
	}

	// Clean up backup dir after successful restore
	if err := os.RemoveAll(BackupDir()); err != nil {
		return actions, fmt.Errorf("cleaning backup dir: %w", err)
	}
	actions = append(actions, "Backup cleared")

	return actions, nil
}

func loadManifest() *Manifest {
	m := &Manifest{Files: make(map[string]FileBackup)}
	data, err := os.ReadFile(manifestPath())
	if err != nil {
		return m
	}
	json.Unmarshal(data, m)
	if m.Files == nil {
		m.Files = make(map[string]FileBackup)
	}
	return m
}

func saveManifest(m *Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath(), data, 0644)
}

// safeFileName converts a path to a safe backup filename
func safeFileName(path string) string {
	// Replace path separators with underscores
	safe := filepath.Base(path)
	dir := filepath.Dir(path)
	dirBase := filepath.Base(dir)
	return dirBase + "_" + safe
}
