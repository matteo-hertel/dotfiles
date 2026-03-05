package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mhdev/dotfiles/tools/colorsync/importer"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "preview",
		Help: "Preview a theme's colors in the terminal",
		Run:  runPreview,
	})
}

func runPreview(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: colorsync preview <theme>")
	}

	theme, err := resolveTheme(args[0])
	if err != nil {
		return err
	}

	preview.Render(os.Stdout, theme)
	return nil
}

func resolveTheme(name string) (*palette.Theme, error) {
	// Try saved themes first
	saved := filepath.Join(palette.ThemesDir(), name+".json")
	if t, err := palette.Load(saved); err == nil {
		return t, nil
	}

	// Try built-in
	if t, err := importer.GetBuiltin(name); err == nil {
		return t, nil
	}

	return nil, fmt.Errorf("theme %q not found (use 'colorsync list' to see available themes)", name)
}
