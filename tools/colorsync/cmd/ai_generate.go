package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/ollama"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "ai-generate",
		Help: "Generate a theme from a natural language description using Ollama",
		Run:  runAIGenerate,
	})
}

func runAIGenerate(args []string) error {
	fs := flag.NewFlagSet("ai-generate", flag.ExitOnError)
	model := fs.String("model", ollama.DefaultModel, "Ollama model to use")
	url := fs.String("url", ollama.DefaultURL, "Ollama API URL")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("usage: colorsync ai-generate [--model qwen3:32b] [--url http://localhost:11434] <description>\n  example: colorsync ai-generate \"a warm dark theme inspired by autumn\"")
	}

	description := strings.Join(remaining, " ")
	fmt.Printf("Generating theme with %s: %q\n", *model, description)
	fmt.Println("This may take a moment...")

	theme, err := ollama.Generate(*url, *model, description)
	if err != nil {
		return err
	}

	if theme.Name == "" {
		return fmt.Errorf("model generated a theme with an empty name")
	}

	preview.Render(os.Stdout, theme)

	reader := bufio.NewReader(os.Stdin)
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
