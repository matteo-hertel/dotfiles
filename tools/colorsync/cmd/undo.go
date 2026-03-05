package cmd

import (
	"fmt"
	"os/exec"

	"github.com/mhdev/dotfiles/tools/colorsync/backup"
)

func init() {
	Register(Command{
		Name: "undo",
		Help: "Undo the last apply, restoring previous state",
		Run:  runUndo,
	})
}

func runUndo(args []string) error {
	// Read manifest before restore (we need metadata for reactivation)
	manifest := backup.GetManifest()
	prevNvim := manifest.NvimPrevColorscheme

	// Restore all files
	actions, err := backup.Restore()
	if err != nil {
		return err
	}

	for _, a := range actions {
		fmt.Println(a)
	}

	// Reactivate previous nvim colorscheme in running instances
	if prevNvim != "" {
		count := sendToRunningNvim(prevNvim)
		if count > 0 {
			fmt.Printf("Neovim: reverted to %q in %d running instance(s)\n", prevNvim, count)
		}
	}

	// Reload tmux to pick up restored config
	if isTmuxRunning() {
		tmuxConf := findTmuxConf()
		if tmuxConf != "" {
			cmd := exec.Command("tmux", "source-file", tmuxConf)
			if err := cmd.Run(); err == nil {
				fmt.Println("tmux: reloaded config")
			}
		}
	}

	fmt.Println("Undo complete.")
	return nil
}
