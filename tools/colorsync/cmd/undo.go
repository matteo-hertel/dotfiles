package cmd

import (
	"fmt"
	"os/exec"

	"github.com/mhdev/dotfiles/tools/colorsync/backup"
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

	manifest := backup.GetManifest()
	prevNvim := manifest.NvimPrevColorscheme

	actions, err := backup.Restore()
	if err != nil {
		return err
	}

	for _, a := range actions {
		fmt.Println(a)
	}

	if prevNvim != "" {
		count := sendToRunningNvim(prevNvim)
		if count > 0 {
			fmt.Printf("Neovim: reverted to %q in %d running instance(s)\n", prevNvim, count)
		}
	}

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
