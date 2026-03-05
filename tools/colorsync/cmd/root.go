package cmd

import (
	"fmt"
	"os"
)

type Command struct {
	Name string
	Help string
	Run  func(args []string) error
}

var commands []Command

func Register(c Command) {
	commands = append(commands, c)
}

func Execute() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	name := os.Args[1]
	for _, c := range commands {
		if c.Name == name {
			if err := c.Run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "unknown command: %s\n", name)
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: colorsync <command> [args]")
	fmt.Fprintln(os.Stderr, "\nCommands:")
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "  %-12s %s\n", c.Name, c.Help)
	}
}
