package preview

import (
	"fmt"
	"io"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func Render(w io.Writer, theme *palette.Theme) {
	fmt.Fprintf(w, "\n  Theme: %s\n\n", theme.Name)

	// Background and foreground swatches
	printSwatch(w, "bg", theme.Background)
	printSwatch(w, "fg", theme.Foreground)
	printSwatch(w, "cursor", theme.Cursor)
	fmt.Fprintln(w)

	// Normal colors (0-7)
	fmt.Fprintf(w, "  Normal:  ")
	for i := 0; i < 8; i++ {
		printBlock(w, theme.Colors[i])
	}
	fmt.Fprintln(w)

	// Bright colors (8-15)
	fmt.Fprintf(w, "  Bright:  ")
	for i := 8; i < 16; i++ {
		printBlock(w, theme.Colors[i])
	}
	fmt.Fprintln(w)

	// Color names row
	names := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	fmt.Fprintf(w, "\n  ")
	for i, name := range names {
		r, g, b, _ := palette.ParseHex(theme.Colors[i])
		fmt.Fprintf(w, "\033[38;2;%d;%d;%dm%-10s\033[0m", r, g, b, name)
	}
	fmt.Fprintln(w)

	// Sample text with foreground on background
	bgR, bgG, bgB, _ := palette.ParseHex(theme.Background)
	fgR, fgG, fgB, _ := palette.ParseHex(theme.Foreground)
	fmt.Fprintf(w, "\n  \033[48;2;%d;%d;%dm\033[38;2;%d;%d;%dm  Sample text on background  \033[0m\n",
		bgR, bgG, bgB, fgR, fgG, fgB)

	// Accent sample using color4 (blue)
	acR, acG, acB, _ := palette.ParseHex(theme.Colors[4])
	fmt.Fprintf(w, "  \033[48;2;%d;%d;%dm\033[38;2;%d;%d;%dm  Accent text on background   \033[0m\n",
		bgR, bgG, bgB, acR, acG, acB)

	fmt.Fprintln(w)
}

func printSwatch(w io.Writer, label, hex string) {
	r, g, b, err := palette.ParseHex(hex)
	if err != nil {
		fmt.Fprintf(w, "  %-8s %s (invalid)\n", label, hex)
		return
	}
	fmt.Fprintf(w, "  %-8s \033[48;2;%d;%d;%dm    \033[0m %s\n", label, r, g, b, hex)
}

func printBlock(w io.Writer, hex string) {
	r, g, b, err := palette.ParseHex(hex)
	if err != nil {
		fmt.Fprintf(w, "  ??  ")
		return
	}
	fmt.Fprintf(w, "\033[48;2;%d;%d;%dm    \033[0m ", r, g, b)
}
