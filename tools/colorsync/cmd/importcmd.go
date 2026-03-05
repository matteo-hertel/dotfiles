package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/importer"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func init() {
	Register(Command{
		Name: "import",
		Help: "Import a theme (built-in name or .itermcolors file)",
		Run:  runImport,
	})
}

func runImport(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: colorsync import <name-or-file>")
	}

	source := args[0]
	var theme *palette.Theme
	var err error

	if strings.HasSuffix(source, ".itermcolors") {
		theme, err = importer.ParseItermColors(source)
	} else {
		theme, err = importer.GetBuiltin(source)
	}
	if err != nil {
		return err
	}

	dir := palette.ThemesDir()
	if err := palette.EnsureDir(dir); err != nil {
		return err
	}

	path := filepath.Join(dir, theme.Name+".json")
	if err := theme.Save(path); err != nil {
		return err
	}

	fmt.Printf("Imported %q -> %s\n", theme.Name, path)
	return nil
}
