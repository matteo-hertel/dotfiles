package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "ai-generate",
		Help: "Generate a theme from a natural language description using Apple Intelligence",
		Run:  runAIGenerate,
	})
}

func runAIGenerate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: colorsync ai-generate <description>\n  example: colorsync ai-generate \"a warm dark theme inspired by autumn\"")
	}

	description := strings.Join(args, " ")
	fmt.Printf("Generating theme: %q\n", description)

	bin, err := findAIBinary()
	if err != nil {
		return err
	}

	cmd := exec.Command(bin, description)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("colorsync-ai failed: %w", err)
	}

	var theme palette.Theme
	if err := json.Unmarshal(out, &theme); err != nil {
		return fmt.Errorf("parsing AI output: %w", err)
	}
	if theme.Name == "" {
		return fmt.Errorf("AI generated a theme with an empty name")
	}

	preview.Render(os.Stdout, &theme)

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

func findAIBinary() (string, error) {
	// Look adjacent to this binary first
	self, err := os.Executable()
	if err == nil {
		adjacent := filepath.Join(filepath.Dir(self), "colorsync-ai")
		if _, err := os.Stat(adjacent); err == nil {
			return adjacent, nil
		}
	}

	// Fall back to PATH
	p, err := exec.LookPath("colorsync-ai")
	if err != nil {
		return "", fmt.Errorf("colorsync-ai not found; build it with: cd tools/colorsync-ai && swift build")
	}
	return p, nil
}
