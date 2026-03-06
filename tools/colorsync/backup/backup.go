package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Manifest tracks one apply's backup state
type Manifest struct {
	Files               map[string]FileBackup `json:"files"`
	NvimPrevColorscheme string                `json:"nvim_prev_colorscheme,omitempty"`
	TmuxSourceAdded     bool                  `json:"tmux_source_added,omitempty"`
	TmuxConfPath        string                `json:"tmux_conf_path,omitempty"`
}

type FileBackup struct {
	BackupPath string `json:"backup_path,omitempty"`
	Existed    bool   `json:"existed"`
}

// Stack holds multiple manifests — one per apply
type Stack struct {
	Entries []Manifest `json:"entries"`
}

func BackupDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "colorsync", "backup")
}

func stackPath() string {
	return filepath.Join(BackupDir(), "stack.json")
}

// BeginApply pushes a new empty manifest onto the stack.
// Call this once at the start of each apply.
func BeginApply() error {
	if err := os.MkdirAll(BackupDir(), 0755); err != nil {
		return err
	}
	s := loadStack()
	s.Entries = append(s.Entries, Manifest{
		Files: make(map[string]FileBackup),
	})
	return saveStack(s)
}

// SaveBackup copies the current file to the backup dir before it gets overwritten.
// Must call BeginApply first.
func SaveBackup(originalPath string) error {
	s := loadStack()
	if len(s.Entries) == 0 {
		return fmt.Errorf("no active apply (call BeginApply first)")
	}
	current := &s.Entries[len(s.Entries)-1]
	idx := len(s.Entries) - 1

	// Create a subdirectory for this stack entry's backup files
	entryDir := filepath.Join(BackupDir(), strconv.Itoa(idx))
	if err := os.MkdirAll(entryDir, 0755); err != nil {
		return err
	}

	info, err := os.Stat(originalPath)
	if os.IsNotExist(err) {
		current.Files[originalPath] = FileBackup{Existed: false}
		return saveStack(s)
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("cannot backup directory: %s", originalPath)
	}

	backupName := safeFileName(originalPath)
	backupPath := filepath.Join(entryDir, backupName)

	data, err := os.ReadFile(originalPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return err
	}

	current.Files[originalPath] = FileBackup{
		BackupPath: backupPath,
		Existed:    true,
	}
	return saveStack(s)
}

// Restore pops the most recent apply and restores its files.
func Restore() ([]string, error) {
	s := loadStack()
	if len(s.Entries) == 0 {
		return nil, fmt.Errorf("nothing to undo (no backup found)")
	}

	// Pop the last entry
	idx := len(s.Entries) - 1
	manifest := s.Entries[idx]
	s.Entries = s.Entries[:idx]

	var actions []string

	for originalPath, info := range manifest.Files {
		if !info.Existed {
			if err := os.Remove(originalPath); err != nil && !os.IsNotExist(err) {
				return actions, fmt.Errorf("removing %s: %w", originalPath, err)
			}
			actions = append(actions, fmt.Sprintf("Removed %s (was newly created)", originalPath))
		} else {
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

	// Clean up this entry's backup files
	entryDir := filepath.Join(BackupDir(), strconv.Itoa(idx))
	os.RemoveAll(entryDir)

	// If stack is now empty, clean up everything
	if len(s.Entries) == 0 {
		os.RemoveAll(BackupDir())
		actions = append(actions, "All backups cleared")
	} else {
		saveStack(s)
		remaining := len(s.Entries)
		actions = append(actions, fmt.Sprintf("%d more undo(s) available", remaining))
	}

	return actions, nil
}

// Depth returns how many undos are available
func Depth() int {
	s := loadStack()
	return len(s.Entries)
}

// ListSnapshots returns all manifests in the stack (oldest first)
func ListSnapshots() []Manifest {
	s := loadStack()
	return s.Entries
}

// SetNvimColorscheme records the previous nvim colorscheme on the current entry
func SetNvimColorscheme(name string) error {
	s := loadStack()
	if len(s.Entries) == 0 {
		return nil
	}
	s.Entries[len(s.Entries)-1].NvimPrevColorscheme = name
	return saveStack(s)
}

// SetTmuxSourceAdded records that we added the source-file line
func SetTmuxSourceAdded(confPath string) error {
	s := loadStack()
	if len(s.Entries) == 0 {
		return nil
	}
	s.Entries[len(s.Entries)-1].TmuxSourceAdded = true
	s.Entries[len(s.Entries)-1].TmuxConfPath = confPath
	return saveStack(s)
}

// GetManifest returns the most recent manifest (for undo logic)
func GetManifest() *Manifest {
	s := loadStack()
	if len(s.Entries) == 0 {
		return &Manifest{Files: make(map[string]FileBackup)}
	}
	return &s.Entries[len(s.Entries)-1]
}

func loadStack() *Stack {
	s := &Stack{}
	data, err := os.ReadFile(stackPath())
	if err != nil {
		return s
	}
	json.Unmarshal(data, s)
	return s
}

func saveStack(s *Stack) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stackPath(), data, 0644)
}

func safeFileName(path string) string {
	safe := filepath.Base(path)
	dir := filepath.Dir(path)
	dirBase := filepath.Base(dir)
	return dirBase + "_" + safe
}
