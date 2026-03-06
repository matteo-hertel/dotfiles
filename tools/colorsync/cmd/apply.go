package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/backup"
	"github.com/mhdev/dotfiles/tools/colorsync/exporter"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "apply",
		Help: "Apply a theme to neovim, tmux, iTerm, and p10k",
		Run:  runApply,
	})
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	targets := fs.String("target", "nvim,tmux,iterm,p10k", "Comma-separated targets: nvim,tmux,iterm,p10k")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		return fmt.Errorf("usage: colorsync apply <theme> [--target nvim,tmux,iterm,p10k]")
	}

	theme, err := resolveTheme(remaining[0])
	if err != nil {
		return err
	}

	preview.Render(os.Stdout, theme)

	// Start a new backup entry for this apply
	if err := backup.BeginApply(); err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	targetSet := make(map[string]bool)
	for _, t := range strings.Split(*targets, ",") {
		targetSet[strings.TrimSpace(t)] = true
	}

	if targetSet["nvim"] {
		if err := applyNeovim(theme); err != nil {
			return fmt.Errorf("neovim: %w", err)
		}
	}

	if targetSet["tmux"] {
		if err := applyTmux(theme); err != nil {
			return fmt.Errorf("tmux: %w", err)
		}
	}

	if targetSet["iterm"] {
		if err := applyIterm(theme); err != nil {
			return fmt.Errorf("iterm: %w", err)
		}
	}

	if targetSet["p10k"] {
		if err := applyP10k(theme); err != nil {
			return fmt.Errorf("p10k: %w", err)
		}
	}

	fmt.Println("\nDone! All targets applied.")
	return nil
}

// --- Neovim ---

func applyNeovim(theme *palette.Theme) error {
	// 1. Write the colorscheme lua file
	path := exporter.NeovimDefaultPath(theme.Name)
	if err := backup.SaveBackup(path); err != nil {
		return fmt.Errorf("backup colorscheme: %w", err)
	}
	if err := exporter.ExportNeovim(theme, path); err != nil {
		return err
	}
	fmt.Printf("Neovim: wrote %s\n", path)

	// 2. Update astroui.lua to set the colorscheme
	astroui := findAstroUI()
	if astroui != "" {
		prevColorscheme := readCurrentColorscheme(astroui)
		if prevColorscheme != "" {
			backup.SetNvimColorscheme(prevColorscheme)
		}
		if err := backup.SaveBackup(astroui); err != nil {
			return fmt.Errorf("backup astroui: %w", err)
		}
		if err := updateAstroUI(astroui, theme.Name); err != nil {
			fmt.Printf("Neovim: warning: could not update astroui.lua: %v\n", err)
		} else {
			fmt.Printf("Neovim: updated %s -> colorscheme = %q\n", astroui, theme.Name)
		}
	}

	// 3. Send :colorscheme to running nvim instances
	count := sendToRunningNvim(theme.Name)
	if count > 0 {
		fmt.Printf("Neovim: applied to %d running instance(s)\n", count)
	}

	return nil
}

func findAstroUI() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "nvim", "lua", "plugins", "astroui.lua")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func readCurrentColorscheme(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`colorscheme\s*=\s*"([^"]+)"`)
	m := re.FindSubmatch(data)
	if len(m) >= 2 {
		return string(m[1])
	}
	return ""
}

func updateAstroUI(path, newScheme string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`(colorscheme\s*=\s*)"[^"]+"`)
	updated := re.ReplaceAll(data, []byte(fmt.Sprintf(`${1}"%s"`, newScheme)))
	return os.WriteFile(path, updated, 0644)
}

func sendToRunningNvim(themeName string) int {
	sockets := findNvimSockets()
	count := 0
	for _, sock := range sockets {
		cmd := exec.Command("nvim", "--server", sock, "--remote-send",
			fmt.Sprintf("<Cmd>colorscheme %s<CR>", themeName))
		if err := cmd.Run(); err == nil {
			count++
		}
	}
	return count
}

func findNvimSockets() []string {
	var sockets []string

	tmpDirs := []string{"/tmp"}
	if tmpdir := os.Getenv("TMPDIR"); tmpdir != "" {
		tmpDirs = append(tmpDirs, tmpdir)
	}

	for _, dir := range tmpDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || !strings.HasPrefix(e.Name(), "nvim") {
				continue
			}
			sock := filepath.Join(dir, e.Name(), "0")
			if info, err := os.Stat(sock); err == nil && info.Mode()&os.ModeSocket != 0 {
				sockets = append(sockets, sock)
			}
		}
	}

	return sockets
}

// --- tmux ---

func applyTmux(theme *palette.Theme) error {
	// 1. Write theme.conf
	path := exporter.TmuxDefaultPath()
	if err := backup.SaveBackup(path); err != nil {
		return fmt.Errorf("backup theme.conf: %w", err)
	}
	if err := exporter.ExportTmux(theme, path); err != nil {
		return err
	}
	fmt.Printf("tmux: wrote %s\n", path)

	// 2. Ensure source-file line in .tmux.conf
	tmuxConf := findTmuxConf()
	if tmuxConf != "" {
		added, err := ensureTmuxSourceLine(tmuxConf, path)
		if err != nil {
			fmt.Printf("tmux: warning: could not update %s: %v\n", tmuxConf, err)
		} else if added {
			backup.SetTmuxSourceAdded(tmuxConf)
			fmt.Printf("tmux: added source-file line to %s\n", tmuxConf)
		}
	}

	// 3. Reload tmux live — source full .tmux.conf so theme.conf overrides
	//    at the end take effect, then force a client refresh
	if isTmuxRunning() {
		if tmuxConf != "" {
			exec.Command("tmux", "source-file", tmuxConf).Run()
		} else {
			exec.Command("tmux", "source-file", path).Run()
		}
		exec.Command("tmux", "refresh-client", "-S").Run()
		fmt.Println("tmux: live reload applied")
	}

	return nil
}

func findTmuxConf() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".tmux.conf")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func ensureTmuxSourceLine(confPath, themePath string) (bool, error) {
	data, err := os.ReadFile(confPath)
	if err != nil {
		return false, err
	}

	sourceLine := fmt.Sprintf("source-file %s", themePath)
	if strings.Contains(string(data), sourceLine) {
		return false, nil // already present
	}

	// Back up .tmux.conf before modifying
	if err := backup.SaveBackup(confPath); err != nil {
		return false, err
	}

	updated := string(data) + "\n# colorsync theme\n" + sourceLine + "\n"
	if err := os.WriteFile(confPath, []byte(updated), 0644); err != nil {
		return false, err
	}
	return true, nil
}

func isTmuxRunning() bool {
	cmd := exec.Command("tmux", "list-sessions")
	return cmd.Run() == nil
}

// --- iTerm ---

func applyIterm(theme *palette.Theme) error {
	// 1. Write .itermcolors file
	filePath := exporter.ItermDefaultPath(theme.Name)
	if err := backup.SaveBackup(filePath); err != nil {
		return fmt.Errorf("backup iterm: %w", err)
	}
	if err := exporter.ExportItermFile(theme, filePath); err != nil {
		return err
	}
	fmt.Printf("iTerm: wrote %s\n", filePath)

	// 2. Live-update running terminal via escape sequences
	//    Enable tmux pass-through if inside tmux so escapes reach iTerm
	if isTmuxRunning() {
		exec.Command("tmux", "set", "-g", "allow-passthrough", "on").Run()
	}
	exporter.WriteItermEscapes(os.Stdout, theme)
	fmt.Println("iTerm: live colors updated")

	return nil
}

// --- p10k ---

func applyP10k(theme *palette.Theme) error {
	path := exporter.P10kDefaultPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("p10k: skipped (no ~/.zshtheme found)")
		return nil
	}
	if err := backup.SaveBackup(path); err != nil {
		return fmt.Errorf("backup p10k: %w", err)
	}
	if err := exporter.ExportP10k(theme, path); err != nil {
		return err
	}
	fmt.Printf("p10k: updated %s\n", path)
	fmt.Println("  Reload with: source ~/.zshtheme")
	return nil
}
