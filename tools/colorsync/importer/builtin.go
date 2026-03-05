package importer

import (
	"fmt"
	"sort"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

var builtins = map[string]palette.Theme{
	"catppuccin-mocha": {
		Name: "catppuccin-mocha", Background: "#1e1e2e", Foreground: "#cdd6f4", Cursor: "#f5e0dc",
		Colors: [16]string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	},
	"catppuccin-latte": {
		Name: "catppuccin-latte", Background: "#eff1f5", Foreground: "#4c4f69", Cursor: "#dc8a78",
		Colors: [16]string{
			"#5c5f77", "#d20f39", "#40a02b", "#df8e1d", "#1e66f5", "#ea76cb", "#179299", "#acb0be",
			"#6c6f85", "#d20f39", "#40a02b", "#df8e1d", "#1e66f5", "#ea76cb", "#179299", "#bcc0cc",
		},
	},
	"gruvbox-dark": {
		Name: "gruvbox-dark", Background: "#282828", Foreground: "#ebdbb2", Cursor: "#ebdbb2",
		Colors: [16]string{
			"#282828", "#cc241d", "#98971a", "#d79921", "#458588", "#b16286", "#689d6a", "#a89984",
			"#928374", "#fb4934", "#b8bb26", "#fabd2f", "#83a598", "#d3869b", "#8ec07c", "#ebdbb2",
		},
	},
	"gruvbox-light": {
		Name: "gruvbox-light", Background: "#fbf1c7", Foreground: "#3c3836", Cursor: "#3c3836",
		Colors: [16]string{
			"#fbf1c7", "#cc241d", "#98971a", "#d79921", "#458588", "#b16286", "#689d6a", "#7c6f64",
			"#928374", "#9d0006", "#79740e", "#b57614", "#076678", "#8f3f71", "#427b58", "#3c3836",
		},
	},
	"tokyo-night": {
		Name: "tokyo-night", Background: "#1a1b26", Foreground: "#c0caf5", Cursor: "#c0caf5",
		Colors: [16]string{
			"#15161e", "#f7768e", "#9ece6a", "#e0af68", "#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6",
			"#414868", "#f7768e", "#9ece6a", "#e0af68", "#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
		},
	},
	"nord": {
		Name: "nord", Background: "#2e3440", Foreground: "#d8dee9", Cursor: "#d8dee9",
		Colors: [16]string{
			"#3b4252", "#bf616a", "#a3be8c", "#ebcb8b", "#81a1c1", "#b48ead", "#88c0d0", "#e5e9f0",
			"#4c566a", "#bf616a", "#a3be8c", "#ebcb8b", "#81a1c1", "#b48ead", "#8fbcbb", "#eceff4",
		},
	},
}

func GetBuiltin(name string) (*palette.Theme, error) {
	t, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("unknown built-in theme: %q (use 'colorsync list' to see available themes)", name)
	}
	copy := t
	return &copy, nil
}

func ListBuiltins() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
