package cmd

import (
	"fmt"

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
	actions, err := backup.Restore()
	if err != nil {
		return err
	}

	for _, a := range actions {
		fmt.Println(a)
	}
	fmt.Println("Undo complete.")
	return nil
}
