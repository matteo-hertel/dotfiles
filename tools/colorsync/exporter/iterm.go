package exporter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

// ExportItermFile writes a valid .itermcolors plist XML file for the given theme.
func ExportItermFile(theme *palette.Theme, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintln(f, `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`)
	fmt.Fprintln(f, `<plist version="1.0">`)
	fmt.Fprintln(f, `<dict>`)

	// ANSI colors 0-15
	for i := 0; i < 16; i++ {
		writeItermColor(f, fmt.Sprintf("Ansi %d Color", i), theme.Colors[i])
	}

	// Special colors
	writeItermColor(f, "Background Color", theme.Background)
	writeItermColor(f, "Foreground Color", theme.Foreground)
	writeItermColor(f, "Cursor Color", theme.Cursor)
	writeItermColor(f, "Cursor Text Color", theme.Background)
	writeItermColor(f, "Selection Color", theme.Colors[8])  // color8
	writeItermColor(f, "Selected Text Color", theme.Foreground)

	fmt.Fprintln(f, `</dict>`)
	fmt.Fprintln(f, `</plist>`)

	return nil
}

// writeItermColor writes a single color entry as a plist XML dict with Alpha=1,
// RGB as float64 (value/255.0), and color space sRGB.
func writeItermColor(w io.Writer, name, hex string) {
	r, g, b, err := palette.ParseHex(hex)
	if err != nil {
		return
	}

	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	fmt.Fprintf(w, "\t<key>%s</key>\n", name)
	fmt.Fprintln(w, "\t<dict>")
	fmt.Fprintln(w, "\t\t<key>Alpha Component</key>")
	fmt.Fprintln(w, "\t\t<real>1</real>")
	fmt.Fprintln(w, "\t\t<key>Blue Component</key>")
	fmt.Fprintf(w, "\t\t<real>%f</real>\n", bf)
	fmt.Fprintln(w, "\t\t<key>Color Space</key>")
	fmt.Fprintln(w, "\t\t<string>sRGB</string>")
	fmt.Fprintln(w, "\t\t<key>Green Component</key>")
	fmt.Fprintf(w, "\t\t<real>%f</real>\n", gf)
	fmt.Fprintln(w, "\t\t<key>Red Component</key>")
	fmt.Fprintf(w, "\t\t<real>%f</real>\n", rf)
	fmt.Fprintln(w, "\t</dict>")
}

// WriteItermEscapes writes iTerm2 proprietary escape sequences for live terminal
// color updates. Format: \033]1337;SetColors=key=rrggbb\007
func WriteItermEscapes(w io.Writer, theme *palette.Theme) {
	ansiKeys := []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"br_black", "br_red", "br_green", "br_yellow", "br_blue", "br_magenta", "br_cyan", "br_white",
	}

	writeEsc := func(key, hex string) {
		hex = strings.TrimPrefix(hex, "#")
		fmt.Fprintf(w, "\033]1337;SetColors=%s=%s\007", key, hex)
	}

	writeEsc("bg", theme.Background)
	writeEsc("fg", theme.Foreground)
	writeEsc("curbg", theme.Cursor)
	writeEsc("curfg", theme.Background)

	for i, key := range ansiKeys {
		writeEsc(key, theme.Colors[i])
	}
}

// ItermDefaultPath returns the default output path for an .itermcolors file.
func ItermDefaultPath(themeName string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "colorsync", "output", themeName+".itermcolors")
}
