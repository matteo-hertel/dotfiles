package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "generate",
		Help: "Generate a theme from bg, fg, and accent colors",
		Run:  runGenerate,
	})
}

func runGenerate(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	bg := prompt(reader, "Background (#hex): ")
	fg := prompt(reader, "Foreground (#hex): ")
	accent := prompt(reader, "Accent (#hex): ")
	name := prompt(reader, "Name: ")

	theme, err := palette.Generate(name, bg, fg, accent)
	if err != nil {
		return err
	}

	preview.Render(os.Stdout, theme)

	if confirm(reader, "Save? [y/n]: ") {
		dir := palette.ThemesDir()
		if err := palette.EnsureDir(dir); err != nil {
			return err
		}
		path := filepath.Join(dir, theme.Name+".json")
		if err := theme.Save(path); err != nil {
			return err
		}
		fmt.Printf("Saved to %s\n", path)
	}

	return nil
}

func prompt(r *bufio.Reader, msg string) string {
	fmt.Print(msg)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func confirm(r *bufio.Reader, msg string) bool {
	answer := prompt(r, msg)
	return strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes"
}
