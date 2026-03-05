package cmd

import (
	"fmt"

	"github.com/mhdev/dotfiles/tools/colorsync/importer"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func init() {
	Register(Command{
		Name: "list",
		Help: "List built-in and saved themes",
		Run:  runList,
	})
}

func runList(args []string) error {
	fmt.Println("Built-in themes:")
	for _, name := range importer.ListBuiltins() {
		fmt.Printf("  %s\n", name)
	}

	dir := palette.ThemesDir()
	themes, err := palette.LoadAll(dir)
	if err == nil && len(themes) > 0 {
		fmt.Println("\nSaved themes:")
		for _, t := range themes {
			fmt.Printf("  %s\n", t.Name)
		}
	}

	return nil
}
