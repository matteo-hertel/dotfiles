package palette

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Theme struct {
	Name       string     `json:"name"`
	Background string     `json:"background"`
	Foreground string     `json:"foreground"`
	Cursor     string     `json:"cursor"`
	Colors     [16]string `json:"colors"`
}

func (t *Theme) Save(path string) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Load(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t Theme
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func LoadAll(dir string) ([]*Theme, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var themes []*Theme
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		t, err := Load(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", e.Name(), err)
		}
		themes = append(themes, t)
	}
	return themes, nil
}

func ParseHex(hex string) (r, g, b uint8, err error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}
	var ri, gi, bi int
	_, err = fmt.Sscanf(hex, "%02x%02x%02x", &ri, &gi, &bi)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}
	return uint8(ri), uint8(gi), uint8(bi), nil
}

func ToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func ThemesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "colorsync", "themes")
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
