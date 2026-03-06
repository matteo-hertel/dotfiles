package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mhdev/dotfiles/tools/colorsync/backup"
	"github.com/mhdev/dotfiles/tools/colorsync/exporter"
)

func init() {
	Register(Command{
		Name: "undo",
		Help: "Undo the last apply (use 'undo list' to see snapshots)",
		Run:  runUndo,
	})
}

func runUndo(args []string) error {
	if len(args) > 0 && args[0] == "list" {
		return runUndoList()
	}

	// Capture manifest data before Restore() pops the entry and
	// potentially deletes the backup directory.
	manifest := backup.GetManifest()
	prevNvim := manifest.NvimPrevColorscheme
	tmuxThemePath := exporter.TmuxDefaultPath()
	tmuxThemeExisted := false
	if info, ok := manifest.Files[tmuxThemePath]; ok {
		tmuxThemeExisted = info.Existed
	}

	actions, err := backup.Restore()
	if err != nil {
		return err
	}

	for _, a := range actions {
		fmt.Println(a)
	}

	// --- Neovim ---
	if prevNvim != "" {
		fmt.Printf("Neovim: colorscheme restored to %q\n", prevNvim)
		count := sendToRunningNvim(prevNvim)
		if count > 0 {
			fmt.Printf("Neovim: applied to %d running instance(s)\n", count)
		}
	}

	// --- tmux ---
	if isTmuxRunning() {
		tmuxConf := findTmuxConf()
		if tmuxThemeExisted {
			// Intermediate undo: theme.conf was restored from backup.
			// Source the main conf (which includes the source-file line)
			// so tmux picks up the restored theme.conf.
			if tmuxConf != "" {
				exec.Command("tmux", "source-file", tmuxConf).Run()
			}
			exec.Command("tmux", "refresh-client", "-S").Run()
			fmt.Println("tmux: reloaded restored theme")
		} else {
			// Base undo: theme.conf was removed (it didn't exist before).
			// The restored .tmux.conf has no source-file line, but the
			// tmux server still has stale theme options in memory.
			// Reset all theme-related options to defaults, then reload.
			resetTmuxTheme()
			if tmuxConf != "" {
				exec.Command("tmux", "source-file", tmuxConf).Run()
			}
			exec.Command("tmux", "refresh-client", "-S").Run()
			fmt.Println("tmux: reset theme to defaults")
		}
	}

	// --- iTerm ---
	// Try to revert terminal colors. If the previous theme can be
	// resolved, send its escape sequences. Otherwise reset to the
	// terminal emulator's default palette.
	if prevNvim != "" {
		if prevTheme, err := resolveTheme(prevNvim); err == nil {
			if isTmuxRunning() {
				exec.Command("tmux", "set", "-g", "allow-passthrough", "on").Run()
			}
			exporter.WriteItermEscapes(os.Stdout, prevTheme)
			fmt.Printf("iTerm: restored %s colors\n", prevNvim)
		} else {
			resetItermColors()
			fmt.Println("iTerm: reset colors to terminal defaults")
		}
	} else {
		resetItermColors()
		fmt.Println("iTerm: reset colors to terminal defaults")
	}

	fmt.Println("\nUndo complete.")
	return nil
}

// resetTmuxTheme sets all colorsync-managed tmux options back to their
// defaults so stale theme colors don't linger after a base undo.
func resetTmuxTheme() {
	defaults := []struct{ opt, val string }{
		{"status-style", "default"},
		{"status-left-style", "default"},
		{"status-left-length", "10"},
		{"status-left", "[#S] "},
		{"status-right-style", "default"},
		{"status-right-length", "40"},
		{"status-right", "\"#22T\" %H:%M %d-%b-%y"},
		{"window-status-format", "#I:#W#F"},
		{"window-status-current-format", "#I:#W#F"},
		{"window-status-style", "default"},
		{"window-status-activity-style", "reverse"},
		{"window-status-separator", "\" \""},
		{"status-justify", "left"},
		{"pane-border-style", "default"},
		{"pane-active-border-style", "#{?pane_in_mode,fg=yellow,#{?synchronize-panes,fg=red,fg=green}}"},
		{"display-panes-colour", "blue"},
		{"display-panes-active-colour", "red"},
		{"clock-mode-colour", "blue"},
		{"clock-mode-style", "24"},
		{"message-style", "fg=black,bg=yellow"},
		{"message-command-style", "fg=yellow,bg=black"},
		{"mode-style", "fg=black,bg=yellow"},
	}

	for _, d := range defaults {
		exec.Command("tmux", "set", "-g", d.opt, d.val).Run()
	}
}

// resetItermColors sends ANSI/xterm reset sequences to restore the
// terminal emulator's default color palette.
func resetItermColors() {
	inTmux := os.Getenv("TMUX") != ""

	writeReset := func(seq string) {
		if inTmux {
			fmt.Fprintf(os.Stdout, "\033Ptmux;\033%s\033\\", seq)
		} else {
			fmt.Fprint(os.Stdout, seq)
		}
	}

	// OSC 110: reset foreground
	writeReset("\033]110\007")
	// OSC 111: reset background
	writeReset("\033]111\007")
	// OSC 112: reset cursor color
	writeReset("\033]112\007")
	// OSC 104: reset all ANSI colors to defaults
	writeReset("\033]104\007")
}

func runUndoList() error {
	depth := backup.Depth()
	if depth == 0 {
		fmt.Println("No snapshots to undo.")
		return nil
	}

	snapshots := backup.ListSnapshots()
	fmt.Printf("%d snapshot(s) available:\n\n", depth)
	for i, snap := range snapshots {
		fmt.Printf("  %d. ", i+1)
		if snap.NvimPrevColorscheme != "" {
			fmt.Printf("previous colorscheme: %s", snap.NvimPrevColorscheme)
		} else {
			fmt.Printf("(initial state)")
		}
		fmt.Println()
		for path, info := range snap.Files {
			if info.Existed {
				fmt.Printf("     restore: %s\n", path)
			} else {
				fmt.Printf("     remove:  %s\n", path)
			}
		}
		fmt.Println()
	}
	return nil
}
