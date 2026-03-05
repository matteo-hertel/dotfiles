package importer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

// ParseItermColors reads an iTerm2 .itermcolors plist file and returns a Theme.
func ParseItermColors(path string) (*palette.Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	keys, dicts := parsePlistTopLevel(string(data))

	colors := make(map[string]string, len(keys))
	for i, key := range keys {
		if i >= len(dicts) {
			break
		}
		r, g, b := parseColorDict(dicts[i])
		hex := palette.ToHex(
			uint8(math.Round(r*255)),
			uint8(math.Round(g*255)),
			uint8(math.Round(b*255)),
		)
		colors[key] = hex
	}

	name := strings.TrimSuffix(filepath.Base(path), ".itermcolors")
	theme := &palette.Theme{Name: name}

	if v, ok := colors["Background Color"]; ok {
		theme.Background = v
	}
	if v, ok := colors["Foreground Color"]; ok {
		theme.Foreground = v
	}
	if v, ok := colors["Cursor Color"]; ok {
		theme.Cursor = v
	}
	for i := 0; i < 16; i++ {
		key := fmt.Sprintf("Ansi %d Color", i)
		if v, ok := colors[key]; ok {
			theme.Colors[i] = v
		}
	}

	return theme, nil
}

// parsePlistTopLevel extracts the alternating key/dict pairs from the
// top-level <dict> in a plist XML document. It returns parallel slices
// of key strings and inner dict XML fragments.
func parsePlistTopLevel(xml string) (keys []string, dicts []string) {
	// Find the outer <dict>...</dict>
	outerStart := strings.Index(xml, "<dict>")
	if outerStart < 0 {
		return nil, nil
	}
	outerEnd := strings.LastIndex(xml, "</dict>")
	if outerEnd < 0 {
		return nil, nil
	}
	inner := xml[outerStart+len("<dict>") : outerEnd]

	// Walk through the inner content finding <key>...</key> and <dict>...</dict> pairs.
	pos := 0
	for {
		// Find next <key>
		keyStart := strings.Index(inner[pos:], "<key>")
		if keyStart < 0 {
			break
		}
		keyStart += pos + len("<key>")
		keyEnd := strings.Index(inner[keyStart:], "</key>")
		if keyEnd < 0 {
			break
		}
		keyEnd += keyStart
		keyVal := strings.TrimSpace(inner[keyStart:keyEnd])
		pos = keyEnd + len("</key>")

		// The next element should be a <dict>
		dictStart := strings.Index(inner[pos:], "<dict>")
		if dictStart < 0 {
			break
		}
		dictStart += pos
		// Find the matching </dict> — inner dicts are flat (no nesting).
		dictEnd := strings.Index(inner[dictStart+len("<dict>"):], "</dict>")
		if dictEnd < 0 {
			break
		}
		dictEnd += dictStart + len("<dict>") + len("</dict>")
		dictContent := inner[dictStart:dictEnd]

		keys = append(keys, keyVal)
		dicts = append(dicts, dictContent)
		pos = dictEnd
	}

	return keys, dicts
}

// parseColorDict extracts the Red, Green, and Blue Component float values
// from an inner color <dict> fragment.
func parseColorDict(dictXML string) (r, g, b float64) {
	components := map[string]*float64{
		"Red Component":   &r,
		"Green Component": &g,
		"Blue Component":  &b,
	}

	for name, ptr := range components {
		idx := strings.Index(dictXML, "<key>"+name+"</key>")
		if idx < 0 {
			continue
		}
		after := dictXML[idx:]
		realStart := strings.Index(after, "<real>")
		if realStart < 0 {
			continue
		}
		realStart += len("<real>")
		realEnd := strings.Index(after[realStart:], "</real>")
		if realEnd < 0 {
			continue
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(after[realStart:realStart+realEnd]), 64)
		if err == nil {
			*ptr = val
		}
	}

	return r, g, b
}
