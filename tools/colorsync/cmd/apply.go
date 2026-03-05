package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/backup"
	"github.com/mhdev/dotfiles/tools/colorsync/exporter"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "apply",
		Help: "Apply a theme to neovim, tmux, and iTerm",
		Run:  runApply,
	})
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	targets := fs.String("target", "nvim,tmux,iterm", "Comma-separated targets: nvim,tmux,iterm")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		return fmt.Errorf("usage: colorsync apply <theme> [--target nvim,tmux,iterm]")
	}

	theme, err := resolveTheme(remaining[0])
	if err != nil {
		return err
	}

	preview.Render(os.Stdout, theme)

	targetSet := make(map[string]bool)
	for _, t := range strings.Split(*targets, ",") {
		targetSet[strings.TrimSpace(t)] = true
	}

	if targetSet["nvim"] {
		path := exporter.NeovimDefaultPath(theme.Name)
		if err := backup.SaveBackup(path); err != nil {
			return fmt.Errorf("backup neovim: %w", err)
		}
		if err := exporter.ExportNeovim(theme, path); err != nil {
			return fmt.Errorf("neovim: %w", err)
		}
		fmt.Printf("Neovim: wrote %s\n", path)
		fmt.Printf("  Activate with: colorscheme %s\n", strings.ReplaceAll(theme.Name, "-", "_"))
	}

	if targetSet["tmux"] {
		path := exporter.TmuxDefaultPath()
		if err := backup.SaveBackup(path); err != nil {
			return fmt.Errorf("backup tmux: %w", err)
		}
		if err := exporter.ExportTmux(theme, path); err != nil {
			return fmt.Errorf("tmux: %w", err)
		}
		fmt.Printf("tmux: wrote %s\n", path)
		fmt.Println("  Add to .tmux.conf: source-file ~/.tmux/theme.conf")
		fmt.Println("  Reload with: tmux source-file ~/.tmux.conf")
	}

	if targetSet["iterm"] {
		filePath := exporter.ItermDefaultPath(theme.Name)
		if err := backup.SaveBackup(filePath); err != nil {
			return fmt.Errorf("backup iterm: %w", err)
		}
		if err := exporter.ExportItermFile(theme, filePath); err != nil {
			return fmt.Errorf("iterm file: %w", err)
		}
		fmt.Printf("iTerm: wrote %s\n", filePath)

		// Live-update running terminal
		exporter.WriteItermEscapes(os.Stdout, theme)
		fmt.Println("iTerm: live colors updated")
	}

	return nil
}
