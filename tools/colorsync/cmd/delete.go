package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func init() {
	Register(Command{
		Name: "delete",
		Help: "Delete one or more saved themes",
		Run:  runDelete,
	})
}

func runDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: colorsync delete <theme-name> [theme-name ...]\n  example: colorsync delete autumn-dusk warm-canopy")
	}

	dir := palette.ThemesDir()
	var deleted, notFound []string

	for _, name := range args {
		path := filepath.Join(dir, name+".json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			notFound = append(notFound, name)
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("deleting %s: %w", name, err)
		}
		deleted = append(deleted, name)
	}

	for _, name := range deleted {
		fmt.Printf("Deleted %s\n", name)
	}
	for _, name := range notFound {
		fmt.Printf("Not found: %s\n", name)
	}

	return nil
}
