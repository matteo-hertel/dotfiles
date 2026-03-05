# colorsync Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool that imports, generates, previews, and applies color schemes across neovim, tmux, and iTerm2.

**Architecture:** Subcommand-based CLI using stdlib only. A unified `Theme` struct (bg, fg, cursor, 16 ANSI colors as hex) is the interchange format. Importers produce themes, exporters consume them. Themes persist as JSON in `~/.config/colorsync/themes/`.

**Tech Stack:** Go stdlib only. No cobra, no third-party deps. `encoding/xml` for itermcolors, `encoding/json` for themes, `os` for file I/O, `fmt` for terminal output with ANSI escapes.

---

### Task 1: Project Scaffold + CLI Dispatch

**Files:**
- Create: `tools/colorsync/go.mod`
- Create: `tools/colorsync/main.go`
- Create: `tools/colorsync/cmd/root.go`

**Step 1: Initialize the Go module**

```bash
mkdir -p tools/colorsync && cd tools/colorsync && go mod init github.com/mhdev/dotfiles/tools/colorsync
```

**Step 2: Create the subcommand dispatcher**

Create `tools/colorsync/cmd/root.go`:

```go
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
```

**Step 3: Create main.go**

Create `tools/colorsync/main.go`:

```go
package main

import "github.com/mhdev/dotfiles/tools/colorsync/cmd"

func main() {
	cmd.Execute()
}
```

**Step 4: Verify it compiles**

```bash
cd tools/colorsync && go build -o colorsync .
./colorsync
```

Expected: prints usage with "Usage: colorsync <command> [args]" and exits non-zero.

**Step 5: Commit**

```bash
git add tools/colorsync/
git commit -m "feat(colorsync): scaffold Go project with CLI dispatch"
```

---

### Task 2: Palette Data Model

**Files:**
- Create: `tools/colorsync/palette/palette.go`
- Create: `tools/colorsync/palette/palette_test.go`

**Step 1: Write the failing test**

Create `tools/colorsync/palette/palette_test.go`:

```go
package palette

import (
	"os"
	"path/filepath"
	"testing"
)

func TestThemeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	theme := Theme{
		Name:       "test-theme",
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Cursor:     "#f5e0dc",
		Colors: [16]string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	}

	path := filepath.Join(dir, "test-theme.json")
	if err := theme.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != theme.Name {
		t.Errorf("Name: got %q, want %q", loaded.Name, theme.Name)
	}
	if loaded.Background != theme.Background {
		t.Errorf("Background: got %q, want %q", loaded.Background, theme.Background)
	}
	if loaded.Colors != theme.Colors {
		t.Errorf("Colors mismatch")
	}
}

func TestLoadThemesDir(t *testing.T) {
	dir := t.TempDir()
	theme := Theme{Name: "alpha", Background: "#000000", Foreground: "#ffffff", Cursor: "#ffffff"}
	theme.Save(filepath.Join(dir, "alpha.json"))

	theme2 := Theme{Name: "beta", Background: "#111111", Foreground: "#eeeeee", Cursor: "#eeeeee"}
	theme2.Save(filepath.Join(dir, "beta.json"))

	themes, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("got %d themes, want 2", len(themes))
	}
}

func TestParseHexColor(t *testing.T) {
	r, g, b, err := ParseHex("#ff8800")
	if err != nil {
		t.Fatalf("ParseHex: %v", err)
	}
	if r != 255 || g != 136 || b != 0 {
		t.Errorf("got (%d,%d,%d), want (255,136,0)", r, g, b)
	}

	_, _, _, err = ParseHex("invalid")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./palette/ -v
```

Expected: compilation failure, `palette` package doesn't exist.

**Step 3: Write the implementation**

Create `tools/colorsync/palette/palette.go`:

```go
package palette

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Theme struct {
	Name       string      `json:"name"`
	Background string      `json:"background"`
	Foreground string      `json:"foreground"`
	Cursor     string      `json:"cursor"`
	Colors     [16]string  `json:"colors"`
}

func (t *Theme) Save(path string) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Load(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t Theme
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func LoadAll(dir string) ([]*Theme, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var themes []*Theme
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		t, err := Load(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", e.Name(), err)
		}
		themes = append(themes, t)
	}
	return themes, nil
}

func ParseHex(hex string) (r, g, b uint8, err error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}
	var ri, gi, bi int
	_, err = fmt.Sscanf(hex, "%02x%02x%02x", &ri, &gi, &bi)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}
	return uint8(ri), uint8(gi), uint8(bi), nil
}

func ToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func ThemesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "colorsync", "themes")
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./palette/ -v
```

Expected: all 3 tests PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/palette/
git commit -m "feat(colorsync): add palette data model with JSON round-trip"
```

---

### Task 3: Built-in Themes

**Files:**
- Create: `tools/colorsync/importer/builtin.go`
- Create: `tools/colorsync/importer/builtin_test.go`

**Step 1: Write the failing test**

Create `tools/colorsync/importer/builtin_test.go`:

```go
package importer

import (
	"testing"
)

func TestGetBuiltin(t *testing.T) {
	theme, err := GetBuiltin("catppuccin-mocha")
	if err != nil {
		t.Fatalf("GetBuiltin: %v", err)
	}
	if theme.Name != "catppuccin-mocha" {
		t.Errorf("Name: got %q", theme.Name)
	}
	if theme.Background != "#1e1e2e" {
		t.Errorf("Background: got %q", theme.Background)
	}
	// All 16 colors should be populated
	for i, c := range theme.Colors {
		if c == "" {
			t.Errorf("color%d is empty", i)
		}
	}
}

func TestGetBuiltinUnknown(t *testing.T) {
	_, err := GetBuiltin("nonexistent")
	if err == nil {
		t.Error("expected error for unknown theme")
	}
}

func TestListBuiltins(t *testing.T) {
	names := ListBuiltins()
	if len(names) < 6 {
		t.Errorf("expected at least 6 builtins, got %d", len(names))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./importer/ -v
```

Expected: compilation failure.

**Step 3: Write the implementation**

Create `tools/colorsync/importer/builtin.go`:

```go
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
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./importer/ -v
```

Expected: all 3 tests PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/importer/
git commit -m "feat(colorsync): add 6 built-in color themes"
```

---

### Task 4: iTerm `.itermcolors` Importer

**Files:**
- Create: `tools/colorsync/importer/itermcolors.go`
- Create: `tools/colorsync/importer/itermcolors_test.go`
- Create: `tools/colorsync/testdata/test.itermcolors`

**Step 1: Create a test fixture**

Create `tools/colorsync/testdata/test.itermcolors` — a minimal valid `.itermcolors` plist with background, foreground, cursor, and ansi 0-15:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Background Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.11764705882352941</real>
		<key>Green Component</key>
		<real>0.11764705882352941</real>
		<key>Blue Component</key>
		<real>0.18039215686274512</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Foreground Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.80392156862745101</real>
		<key>Green Component</key>
		<real>0.83921568627450982</real>
		<key>Blue Component</key>
		<real>0.95686274509803926</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Cursor Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.96078431372549022</real>
		<key>Green Component</key>
		<real>0.87843137254901960</real>
		<key>Blue Component</key>
		<real>0.86274509803921573</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 0 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.27058823529411763</real>
		<key>Green Component</key>
		<real>0.27843137254901962</real>
		<key>Blue Component</key>
		<real>0.35294117647058826</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 1 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.95294117647058818</real>
		<key>Green Component</key>
		<real>0.54509803921568623</real>
		<key>Blue Component</key>
		<real>0.65882352941176470</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 2 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.65098039215686276</real>
		<key>Green Component</key>
		<real>0.89019607843137236</real>
		<key>Blue Component</key>
		<real>0.63137254901960782</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 3 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.97647058823529409</real>
		<key>Green Component</key>
		<real>0.88627450980392153</real>
		<key>Blue Component</key>
		<real>0.68627450980392157</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 4 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.53725490196078429</real>
		<key>Green Component</key>
		<real>0.70588235294117652</real>
		<key>Blue Component</key>
		<real>0.98039215686274506</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 5 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.96078431372549022</real>
		<key>Green Component</key>
		<real>0.76078431372549016</real>
		<key>Blue Component</key>
		<real>0.90588235294117647</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 6 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.58039215686274510</real>
		<key>Green Component</key>
		<real>0.88627450980392153</real>
		<key>Blue Component</key>
		<real>0.83529411764705885</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 7 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.72941176470588232</real>
		<key>Green Component</key>
		<real>0.76078431372549016</real>
		<key>Blue Component</key>
		<real>0.87058823529411766</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 8 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.34509803921568627</real>
		<key>Green Component</key>
		<real>0.35686274509803922</real>
		<key>Blue Component</key>
		<real>0.43921568627450980</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 9 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.95294117647058818</real>
		<key>Green Component</key>
		<real>0.54509803921568623</real>
		<key>Blue Component</key>
		<real>0.65882352941176470</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 10 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.65098039215686276</real>
		<key>Green Component</key>
		<real>0.89019607843137236</real>
		<key>Blue Component</key>
		<real>0.63137254901960782</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 11 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.97647058823529409</real>
		<key>Green Component</key>
		<real>0.88627450980392153</real>
		<key>Blue Component</key>
		<real>0.68627450980392157</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 12 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.53725490196078429</real>
		<key>Green Component</key>
		<real>0.70588235294117652</real>
		<key>Blue Component</key>
		<real>0.98039215686274506</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 13 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.96078431372549022</real>
		<key>Green Component</key>
		<real>0.76078431372549016</real>
		<key>Blue Component</key>
		<real>0.90588235294117647</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 14 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.58039215686274510</real>
		<key>Green Component</key>
		<real>0.88627450980392153</real>
		<key>Blue Component</key>
		<real>0.83529411764705885</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
	<key>Ansi 15 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.65098039215686276</real>
		<key>Green Component</key>
		<real>0.67843137254901964</real>
		<key>Blue Component</key>
		<real>0.78431372549019607</real>
		<key>Alpha Component</key>
		<real>1</real>
		<key>Color Space</key>
		<string>sRGB</string>
	</dict>
</dict>
</plist>
```

**Step 2: Write the failing test**

Create `tools/colorsync/importer/itermcolors_test.go`:

```go
package importer

import (
	"testing"
)

func TestParseItermColors(t *testing.T) {
	theme, err := ParseItermColors("../testdata/test.itermcolors")
	if err != nil {
		t.Fatalf("ParseItermColors: %v", err)
	}

	if theme.Background != "#1e1e2e" {
		t.Errorf("Background: got %q, want %q", theme.Background, "#1e1e2e")
	}
	if theme.Foreground != "#cdd6f4" {
		t.Errorf("Foreground: got %q, want %q", theme.Foreground, "#cdd6f4")
	}
	if theme.Cursor != "#f5e0dc" {
		t.Errorf("Cursor: got %q, want %q", theme.Cursor, "#f5e0dc")
	}
	if theme.Colors[0] != "#45475a" {
		t.Errorf("Color0: got %q, want %q", theme.Colors[0], "#45475a")
	}
	if theme.Colors[4] != "#89b4fa" {
		t.Errorf("Color4: got %q, want %q", theme.Colors[4], "#89b4fa")
	}
}
```

**Step 3: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./importer/ -v
```

Expected: FAIL, `ParseItermColors` not defined.

**Step 4: Write the implementation**

Create `tools/colorsync/importer/itermcolors.go`:

```go
package importer

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

type plistDict struct {
	XMLName xml.Name `xml:"dict"`
	Content []byte   `xml:",innerxml"`
}

func ParseItermColors(path string) (*palette.Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	colors := make(map[string]string)
	keys, dicts := parsePlistTopLevel(data)

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

func parsePlistTopLevel(data []byte) (keys []string, dicts []string) {
	s := string(data)

	// Find the outer <dict> content
	start := strings.Index(s, "<dict>")
	end := strings.LastIndex(s, "</dict>")
	if start < 0 || end < 0 {
		return nil, nil
	}
	inner := s[start+6 : end]

	// Parse alternating <key> and <dict> elements
	for {
		ki := strings.Index(inner, "<key>")
		if ki < 0 {
			break
		}
		ke := strings.Index(inner[ki:], "</key>")
		if ke < 0 {
			break
		}
		key := inner[ki+5 : ki+ke]

		inner = inner[ki+ke+6:]

		di := strings.Index(inner, "<dict>")
		if di < 0 {
			break
		}
		de := strings.Index(inner[di:], "</dict>")
		if de < 0 {
			break
		}
		dict := inner[di+6 : di+de]
		inner = inner[di+de+7:]

		keys = append(keys, key)
		dicts = append(dicts, dict)
	}

	return keys, dicts
}

func parseColorDict(dict string) (r, g, b float64) {
	components := make(map[string]float64)
	s := dict

	for {
		ki := strings.Index(s, "<key>")
		if ki < 0 {
			break
		}
		ke := strings.Index(s[ki:], "</key>")
		if ke < 0 {
			break
		}
		key := s[ki+5 : ki+ke]
		s = s[ki+ke+6:]

		ri := strings.Index(s, "<real>")
		si := strings.Index(s, "<string>")

		if ri >= 0 && (si < 0 || ri < si) {
			re := strings.Index(s[ri:], "</real>")
			if re >= 0 {
				var val float64
				fmt.Sscanf(s[ri+6:ri+re], "%f", &val)
				components[key] = val
			}
			if re >= 0 {
				s = s[ri+re+7:]
			}
		} else if si >= 0 {
			se := strings.Index(s[si:], "</string>")
			if se >= 0 {
				s = s[si+se+9:]
			}
		}
	}

	return components["Red Component"], components["Green Component"], components["Blue Component"]
}
```

**Step 5: Run tests**

```bash
cd tools/colorsync && go test ./importer/ -v
```

Expected: all tests PASS.

**Step 6: Commit**

```bash
git add tools/colorsync/importer/ tools/colorsync/testdata/
git commit -m "feat(colorsync): add iTerm .itermcolors XML importer"
```

---

### Task 5: Color Generation from bg/fg/accent

**Files:**
- Create: `tools/colorsync/palette/generate.go`
- Create: `tools/colorsync/palette/generate_test.go`

**Step 1: Write the failing test**

Create `tools/colorsync/palette/generate_test.go`:

```go
package palette

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	theme, err := Generate("test-gen", "#1a1b26", "#c0caf5", "#7aa2f7")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if theme.Name != "test-gen" {
		t.Errorf("Name: got %q", theme.Name)
	}
	if theme.Background != "#1a1b26" {
		t.Errorf("Background: got %q", theme.Background)
	}
	if theme.Foreground != "#c0caf5" {
		t.Errorf("Foreground: got %q", theme.Foreground)
	}

	// All 16 colors should be valid hex
	for i, c := range theme.Colors {
		if len(c) != 7 || c[0] != '#' {
			t.Errorf("color%d invalid hex: %q", i, c)
		}
	}

	// color1 (red) should have high red component
	r, _, _, _ := ParseHex(theme.Colors[1])
	if r < 150 {
		t.Errorf("color1 (red) should have high red, got R=%d", r)
	}

	// Bright variants (8-15) should be brighter than normal (0-7)
	for i := 0; i < 8; i++ {
		_, _, _, err1 := ParseHex(theme.Colors[i])
		_, _, _, err2 := ParseHex(theme.Colors[i+8])
		if err1 != nil || err2 != nil {
			t.Errorf("invalid color at %d or %d", i, i+8)
		}
	}
}

func TestGenerateInvalidHex(t *testing.T) {
	_, err := Generate("bad", "invalid", "#ffffff", "#ffffff")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./palette/ -v -run TestGenerate
```

Expected: FAIL, `Generate` not defined.

**Step 3: Write the implementation**

Create `tools/colorsync/palette/generate.go`:

```go
package palette

import (
	"math"
)

// Generate creates a full 16-color theme from background, foreground, and accent colors.
// It derives ANSI colors by distributing hues and adjusting lightness.
func Generate(name, bg, fg, accent string) (*Theme, error) {
	bgR, bgG, bgB, err := ParseHex(bg)
	if err != nil {
		return nil, err
	}
	fgR, fgG, fgB, err := ParseHex(fg)
	if err != nil {
		return nil, err
	}
	acR, acG, acB, err := ParseHex(accent)
	if err != nil {
		return nil, err
	}

	_, bgS, bgL := rgbToHSL(bgR, bgG, bgB)
	_, _, fgL := rgbToHSL(fgR, fgG, fgB)
	acH, acS, _ := rgbToHSL(acR, acG, acB)

	isDark := bgL < 0.5

	// Derive saturation and lightness for generated colors
	sat := acS
	if sat < 0.3 {
		sat = 0.5
	}

	var normalL, brightL float64
	if isDark {
		normalL = 0.55
		brightL = 0.70
	} else {
		normalL = 0.40
		brightL = 0.30
	}

	// 6 hues evenly offset from accent, covering red, green, yellow, blue, magenta, cyan
	hues := [6]float64{
		0,    // red
		120,  // green
		60,   // yellow
		acH,  // blue (use accent hue)
		300,  // magenta
		180,  // cyan
	}

	// If accent is already blue-ish, shift blue slot slightly
	if acH > 200 && acH < 260 {
		hues[3] = acH
	} else {
		hues[3] = 220
	}

	theme := &Theme{
		Name:       name,
		Background: bg,
		Foreground: fg,
		Cursor:     fg,
	}

	// color0: dark shade (slightly lighter than bg for dark themes, slightly darker for light)
	if isDark {
		theme.Colors[0] = hslToHex(0, bgS, clamp(bgL+0.05))
	} else {
		theme.Colors[0] = hslToHex(0, bgS, clamp(bgL-0.05))
	}

	// colors 1-6: the six hues at normal lightness
	for i, h := range hues {
		theme.Colors[i+1] = hslToHex(h, sat, normalL)
	}

	// color7: light foreground variant
	if isDark {
		theme.Colors[7] = hslToHex(0, 0.05, clamp(fgL-0.1))
	} else {
		theme.Colors[7] = hslToHex(0, 0.05, clamp(fgL+0.1))
	}

	// color8: brighter than color0
	if isDark {
		theme.Colors[8] = hslToHex(0, bgS, clamp(bgL+0.15))
	} else {
		theme.Colors[8] = hslToHex(0, bgS, clamp(bgL-0.15))
	}

	// colors 9-14: the six hues at bright lightness
	for i, h := range hues {
		theme.Colors[i+9] = hslToHex(h, sat, brightL)
	}

	// color15: near-foreground
	theme.Colors[15] = fg

	return theme, nil
}

func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l = (max + min) / 2.0

	if max == min {
		return 0, 0, l
	}

	d := max - min
	if l > 0.5 {
		s = d / (2.0 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	case bf:
		h = (rf-gf)/d + 4
	}
	h *= 60

	return h, s, l
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(math.Round(l * 255))
		return v, v, v
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	hNorm := h / 360.0
	r := hueToRGB(p, q, hNorm+1.0/3.0)
	g := hueToRGB(p, q, hNorm)
	b := hueToRGB(p, q, hNorm-1.0/3.0)

	return uint8(math.Round(r * 255)), uint8(math.Round(g * 255)), uint8(math.Round(b * 255))
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func hslToHex(h, s, l float64) string {
	r, g, b := hslToRGB(h, s, l)
	return ToHex(r, g, b)
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./palette/ -v
```

Expected: all tests PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/palette/generate.go tools/colorsync/palette/generate_test.go
git commit -m "feat(colorsync): add palette generation from bg/fg/accent"
```

---

### Task 6: Terminal Preview

**Files:**
- Create: `tools/colorsync/preview/preview.go`
- Create: `tools/colorsync/preview/preview_test.go`

**Step 1: Write the failing test**

Create `tools/colorsync/preview/preview_test.go`:

```go
package preview

import (
	"bytes"
	"testing"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func TestRender(t *testing.T) {
	theme := &palette.Theme{
		Name: "test", Background: "#1e1e2e", Foreground: "#cdd6f4", Cursor: "#f5e0dc",
		Colors: [16]string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	}

	var buf bytes.Buffer
	Render(&buf, theme)
	out := buf.String()

	if len(out) == 0 {
		t.Error("Render produced empty output")
	}
	// Should contain the theme name
	if !bytes.Contains([]byte(out), []byte("test")) {
		t.Error("output should contain theme name")
	}
	// Should contain ANSI escape sequences
	if !bytes.Contains([]byte(out), []byte("\033[")) {
		t.Error("output should contain ANSI escape sequences")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./preview/ -v
```

Expected: FAIL, package doesn't exist.

**Step 3: Write the implementation**

Create `tools/colorsync/preview/preview.go`:

```go
package preview

import (
	"fmt"
	"io"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func Render(w io.Writer, theme *palette.Theme) {
	fmt.Fprintf(w, "\n  Theme: %s\n\n", theme.Name)

	// Background and foreground
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

	// Color names
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
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./preview/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/preview/
git commit -m "feat(colorsync): add terminal color swatch preview"
```

---

### Task 7: Neovim Lua Exporter

**Files:**
- Create: `tools/colorsync/exporter/neovim.go`
- Create: `tools/colorsync/exporter/neovim_test.go`

**Step 1: Write the failing test**

Create `tools/colorsync/exporter/neovim_test.go`:

```go
package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func testTheme() *palette.Theme {
	return &palette.Theme{
		Name: "test-theme", Background: "#1e1e2e", Foreground: "#cdd6f4", Cursor: "#f5e0dc",
		Colors: [16]string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	}
}

func TestExportNeovim(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-theme.lua")

	err := ExportNeovim(testTheme(), path)
	if err != nil {
		t.Fatalf("ExportNeovim: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	// Should set background
	if !strings.Contains(content, "vim.o.background") {
		t.Error("missing vim.o.background")
	}
	// Should clear highlights
	if !strings.Contains(content, "hi clear") {
		t.Error("missing hi clear")
	}
	// Should set Normal group
	if !strings.Contains(content, "Normal") {
		t.Error("missing Normal highlight group")
	}
	// Should reference our colors
	if !strings.Contains(content, "#1e1e2e") {
		t.Error("missing background color")
	}
	if !strings.Contains(content, "#cdd6f4") {
		t.Error("missing foreground color")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./exporter/ -v
```

Expected: FAIL.

**Step 3: Write the implementation**

Create `tools/colorsync/exporter/neovim.go`:

```go
package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

const nvimTemplate = `-- Generated by colorsync - do not edit manually
vim.cmd("hi clear")
if vim.fn.exists("syntax_on") then vim.cmd("syntax reset") end
vim.o.background = "{{ .Background }}"
vim.g.colors_name = "{{ .Name }}"

local c = {
  bg       = "{{ .Theme.Background }}",
  fg       = "{{ .Theme.Foreground }}",
  cursor   = "{{ .Theme.Cursor }}",
  black    = "{{ index .Theme.Colors 0 }}",
  red      = "{{ index .Theme.Colors 1 }}",
  green    = "{{ index .Theme.Colors 2 }}",
  yellow   = "{{ index .Theme.Colors 3 }}",
  blue     = "{{ index .Theme.Colors 4 }}",
  magenta  = "{{ index .Theme.Colors 5 }}",
  cyan     = "{{ index .Theme.Colors 6 }}",
  white    = "{{ index .Theme.Colors 7 }}",
  br_black   = "{{ index .Theme.Colors 8 }}",
  br_red     = "{{ index .Theme.Colors 9 }}",
  br_green   = "{{ index .Theme.Colors 10 }}",
  br_yellow  = "{{ index .Theme.Colors 11 }}",
  br_blue    = "{{ index .Theme.Colors 12 }}",
  br_magenta = "{{ index .Theme.Colors 13 }}",
  br_cyan    = "{{ index .Theme.Colors 14 }}",
  br_white   = "{{ index .Theme.Colors 15 }}",
}

local hi = function(group, opts)
  vim.api.nvim_set_hl(0, group, opts)
end

-- Editor
hi("Normal",       { fg = c.fg, bg = c.bg })
hi("NormalFloat",  { fg = c.fg, bg = c.black })
hi("CursorLine",   { bg = c.black })
hi("CursorColumn", { bg = c.black })
hi("ColorColumn",  { bg = c.black })
hi("LineNr",       { fg = c.br_black })
hi("CursorLineNr", { fg = c.yellow, bold = true })
hi("SignColumn",   { bg = c.bg })
hi("VertSplit",    { fg = c.br_black })
hi("StatusLine",   { fg = c.fg, bg = c.black })
hi("StatusLineNC", { fg = c.br_black, bg = c.black })
hi("Pmenu",        { fg = c.fg, bg = c.black })
hi("PmenuSel",     { fg = c.bg, bg = c.blue })
hi("Visual",       { bg = c.br_black })
hi("Search",       { fg = c.bg, bg = c.yellow })
hi("IncSearch",    { fg = c.bg, bg = c.yellow, bold = true })
hi("MatchParen",   { fg = c.yellow, bold = true })
hi("Folded",       { fg = c.br_black, bg = c.black })
hi("NonText",      { fg = c.br_black })

-- Syntax
hi("Comment",     { fg = c.br_black, italic = true })
hi("Constant",    { fg = c.yellow })
hi("String",      { fg = c.green })
hi("Character",   { fg = c.green })
hi("Number",      { fg = c.yellow })
hi("Boolean",     { fg = c.yellow })
hi("Float",       { fg = c.yellow })
hi("Identifier",  { fg = c.fg })
hi("Function",    { fg = c.blue })
hi("Statement",   { fg = c.magenta })
hi("Conditional", { fg = c.magenta })
hi("Repeat",      { fg = c.magenta })
hi("Label",       { fg = c.magenta })
hi("Operator",    { fg = c.cyan })
hi("Keyword",     { fg = c.magenta })
hi("Exception",   { fg = c.red })
hi("PreProc",     { fg = c.cyan })
hi("Include",     { fg = c.blue })
hi("Define",      { fg = c.magenta })
hi("Type",        { fg = c.cyan })
hi("StorageClass",{ fg = c.yellow })
hi("Structure",   { fg = c.cyan })
hi("Special",     { fg = c.cyan })
hi("SpecialChar", { fg = c.yellow })
hi("Error",       { fg = c.red, bold = true })
hi("Todo",        { fg = c.yellow, bold = true })
hi("Underlined",  { fg = c.blue, underline = true })

-- Diagnostics
hi("DiagnosticError", { fg = c.red })
hi("DiagnosticWarn",  { fg = c.yellow })
hi("DiagnosticInfo",  { fg = c.blue })
hi("DiagnosticHint",  { fg = c.cyan })

-- Git
hi("DiffAdd",    { fg = c.green })
hi("DiffChange", { fg = c.yellow })
hi("DiffDelete", { fg = c.red })
hi("DiffText",   { fg = c.blue })

-- Treesitter
hi("@variable",        { fg = c.fg })
hi("@function",        { fg = c.blue })
hi("@function.call",   { fg = c.blue })
hi("@method",          { fg = c.blue })
hi("@keyword",         { fg = c.magenta })
hi("@string",          { fg = c.green })
hi("@comment",         { fg = c.br_black, italic = true })
hi("@type",            { fg = c.cyan })
hi("@property",        { fg = c.fg })
hi("@parameter",       { fg = c.fg })
hi("@punctuation",     { fg = c.br_black })
hi("@tag",             { fg = c.red })
hi("@tag.attribute",   { fg = c.yellow })
hi("@tag.delimiter",   { fg = c.br_black })
hi("@constructor",     { fg = c.cyan })
hi("@constant.builtin",{ fg = c.yellow })
`

type nvimData struct {
	Name       string
	Background string
	Theme      *palette.Theme
}

func ExportNeovim(theme *palette.Theme, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	_, _, _, err := palette.ParseHex(theme.Background)
	if err != nil {
		return err
	}

	bgMode := "dark"
	r, g, b, _ := palette.ParseHex(theme.Background)
	lum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0
	if lum > 0.5 {
		bgMode = "light"
	}

	// Sanitize name for Lua (replace hyphens with underscores)
	luaName := strings.ReplaceAll(theme.Name, "-", "_")

	tmpl, err := template.New("nvim").Parse(nvimTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, nvimData{
		Name:       luaName,
		Background: bgMode,
		Theme:      theme,
	})
}

func NeovimDefaultPath(themeName string) string {
	home, _ := os.UserHomeDir()
	name := strings.ReplaceAll(themeName, "-", "_")
	return filepath.Join(home, ".config", "nvim", "colors", name+".lua")
}

func FormatNeovimActivation(themeName string) string {
	name := strings.ReplaceAll(themeName, "-", "_")
	return fmt.Sprintf("vim.cmd('colorscheme %s')", name)
}
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./exporter/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/exporter/
git commit -m "feat(colorsync): add neovim Lua colorscheme exporter"
```

---

### Task 8: tmux Exporter

**Files:**
- Create: `tools/colorsync/exporter/tmux.go`
- Create: `tools/colorsync/exporter/tmux_test.go`

**Step 1: Write the failing test**

Create `tools/colorsync/exporter/tmux_test.go`:

```go
package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportTmux(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.conf")

	err := ExportTmux(testTheme(), path)
	if err != nil {
		t.Fatalf("ExportTmux: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "status-style") {
		t.Error("missing status-style")
	}
	if !strings.Contains(content, "pane-border-style") {
		t.Error("missing pane-border-style")
	}
	if !strings.Contains(content, "#1e1e2e") {
		t.Error("missing background color")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./exporter/ -v -run TestExportTmux
```

Expected: FAIL, `ExportTmux` not defined.

**Step 3: Write the implementation**

Create `tools/colorsync/exporter/tmux.go`:

```go
package exporter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func ExportTmux(theme *palette.Theme, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	bg := theme.Background
	fg := theme.Foreground
	accent := theme.Colors[4]  // blue
	black := theme.Colors[0]
	brBlack := theme.Colors[8]
	white := theme.Colors[7]

	conf := fmt.Sprintf(`# Generated by colorsync - do not edit manually
# Theme: %s

# Basic status bar
set -g status-style "fg=%s,bg=%s"

# Left side of status bar
set -g status-left-style "fg=%s,bg=%s"
set -g status-left-length 40
set -g status-left "#[fg=%s,bg=%s,bold] #S #[fg=%s,bg=%s,nobold]#[fg=%s,bg=%s] #(whoami) #[fg=%s,bg=%s]#[fg=%s,bg=%s] #I:#P #[fg=%s,bg=%s,nobold]"

# Right side of status bar
set -g status-right-style "fg=%s,bg=%s"
set -g status-right-length 150
set -g status-right "#[fg=%s,bg=%s]#[fg=%s,bg=%s] %%H:%%M:%%S #[fg=%s,bg=%s]#[fg=%s,bg=%s] %%d/%%B/%%Y #[fg=%s,bg=%s]#[fg=%s,bg=%s,bold] #H "

# Window status
set -g window-status-format "  #I:#W#F  "
set -g window-status-current-format "#[fg=%s,bg=%s]#[fg=%s,nobold] #I:#W#F #[fg=%s,bg=%s,nobold]"

# Current window status
set -g window-status-style "fg=%s,bg=%s"

# Window with activity status
set -g window-status-activity-style "fg=%s,bg=%s"

# Window separator
set -g window-status-separator ""

# Window status alignment
set -g status-justify centre

# Pane border
set -g pane-border-style "fg=%s,bg=default"

# Active pane border
set -g pane-active-border-style "fg=%s,bg=default"

# Pane number indicator
set -g display-panes-colour "%s"
set -g display-panes-active-colour "%s"

# Clock mode
set -g clock-mode-colour "%s"
set -g clock-mode-style 24

# Message
set -g message-style "fg=%s,bg=%s"

# Command message
set -g message-command-style "fg=%s,bg=%s"

# Mode
set -g mode-style "fg=%s,bg=%s"
`,
		theme.Name,
		// status bar
		brBlack, bg,
		// left style
		white, bg,
		// left format
		bg, accent, accent, brBlack, bg, brBlack, brBlack, black, black, brBlack, black, bg,
		// right style
		white, bg,
		// right format
		black, bg, brBlack, black, brBlack, black, bg, brBlack, accent, brBlack, bg, accent,
		// window current
		bg, black, accent, bg, black,
		// window style
		black, accent,
		// activity
		bg, white,
		// pane border
		brBlack,
		// active pane
		accent,
		// pane numbers
		bg, white,
		// clock
		accent,
		// message
		black, accent,
		// command message
		black, bg,
		// mode
		bg, accent,
	)

	return os.WriteFile(path, []byte(conf), 0644)
}

func TmuxDefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tmux", "theme.conf")
}
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./exporter/ -v
```

Expected: all exporter tests PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/exporter/tmux.go tools/colorsync/exporter/tmux_test.go
git commit -m "feat(colorsync): add tmux theme.conf exporter"
```

---

### Task 9: iTerm Exporter

**Files:**
- Create: `tools/colorsync/exporter/iterm.go`
- Create: `tools/colorsync/exporter/iterm_test.go`

**Step 1: Write the failing test**

Create `tools/colorsync/exporter/iterm_test.go`:

```go
package exporter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportItermFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.itermcolors")

	err := ExportItermFile(testTheme(), path)
	if err != nil {
		t.Fatalf("ExportItermFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Background Color") {
		t.Error("missing Background Color")
	}
	if !strings.Contains(content, "Foreground Color") {
		t.Error("missing Foreground Color")
	}
	if !strings.Contains(content, "Ansi 0 Color") {
		t.Error("missing Ansi 0 Color")
	}
	if !strings.Contains(content, "plist") {
		t.Error("missing plist header")
	}
}

func TestItermEscapeSequences(t *testing.T) {
	var buf bytes.Buffer
	WriteItermEscapes(&buf, testTheme())
	out := buf.String()

	if len(out) == 0 {
		t.Error("empty escape sequence output")
	}
	// Should contain OSC sequences
	if !strings.Contains(out, "\033]") {
		t.Error("missing OSC escape sequences")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./exporter/ -v -run TestExportIterm
```

Expected: FAIL.

**Step 3: Write the implementation**

Create `tools/colorsync/exporter/iterm.go`:

```go
package exporter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func ExportItermFile(theme *palette.Theme, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprint(f, `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
`)

	writeItermColor(f, "Background Color", theme.Background)
	writeItermColor(f, "Foreground Color", theme.Foreground)
	writeItermColor(f, "Cursor Color", theme.Cursor)
	writeItermColor(f, "Cursor Text Color", theme.Background)
	writeItermColor(f, "Selection Color", theme.Colors[8])
	writeItermColor(f, "Selected Text Color", theme.Foreground)

	for i := 0; i < 16; i++ {
		writeItermColor(f, fmt.Sprintf("Ansi %d Color", i), theme.Colors[i])
	}

	fmt.Fprint(f, `</dict>
</plist>
`)
	return nil
}

func writeItermColor(w io.Writer, name, hex string) {
	r, g, b, err := palette.ParseHex(hex)
	if err != nil {
		return
	}
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	fmt.Fprintf(w, "\t<key>%s</key>\n", name)
	fmt.Fprintf(w, "\t<dict>\n")
	fmt.Fprintf(w, "\t\t<key>Alpha Component</key>\n\t\t<real>1</real>\n")
	fmt.Fprintf(w, "\t\t<key>Blue Component</key>\n\t\t<real>%.17f</real>\n", bf)
	fmt.Fprintf(w, "\t\t<key>Color Space</key>\n\t\t<string>sRGB</string>\n")
	fmt.Fprintf(w, "\t\t<key>Green Component</key>\n\t\t<real>%.17f</real>\n", gf)
	fmt.Fprintf(w, "\t\t<key>Red Component</key>\n\t\t<real>%.17f</real>\n", rf)
	fmt.Fprintf(w, "\t</dict>\n")
}

// WriteItermEscapes writes proprietary iTerm2 escape sequences to live-update terminal colors.
func WriteItermEscapes(w io.Writer, theme *palette.Theme) {
	// iTerm2 proprietary escape: \033]1337;SetColors=key=rrggbb\007
	writeItermEsc(w, "bg", theme.Background)
	writeItermEsc(w, "fg", theme.Foreground)
	writeItermEsc(w, "curbg", theme.Cursor)
	writeItermEsc(w, "curfg", theme.Background)

	ansiNames := []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"br_black", "br_red", "br_green", "br_yellow", "br_blue", "br_magenta", "br_cyan", "br_white",
	}
	for i, name := range ansiNames {
		writeItermEsc(w, name, theme.Colors[i])
	}
}

func writeItermEsc(w io.Writer, key, hex string) {
	r, g, b, err := palette.ParseHex(hex)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "\033]1337;SetColors=%s=%02x%02x%02x\007", key, r, g, b)
}

func ItermDefaultPath(themeName string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "colorsync", "output", themeName+".itermcolors")
}
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./exporter/ -v
```

Expected: all tests PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/exporter/iterm.go tools/colorsync/exporter/iterm_test.go
git commit -m "feat(colorsync): add iTerm exporter with escape sequences and .itermcolors file"
```

---

### Task 10: CLI Commands

**Files:**
- Create: `tools/colorsync/cmd/list.go`
- Create: `tools/colorsync/cmd/importcmd.go`
- Create: `tools/colorsync/cmd/generate.go`
- Create: `tools/colorsync/cmd/preview.go`
- Create: `tools/colorsync/cmd/apply.go`
- Modify: `tools/colorsync/main.go`

**Step 1: Create list command**

Create `tools/colorsync/cmd/list.go`:

```go
package cmd

import (
	"fmt"

	"github.com/mhdev/dotfiles/tools/colorsync/importer"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func init() {
	Register(Command{
		Name: "list",
		Help: "List built-in and saved themes",
		Run:  runList,
	})
}

func runList(args []string) error {
	fmt.Println("Built-in themes:")
	for _, name := range importer.ListBuiltins() {
		fmt.Printf("  %s\n", name)
	}

	dir := palette.ThemesDir()
	themes, err := palette.LoadAll(dir)
	if err == nil && len(themes) > 0 {
		fmt.Println("\nSaved themes:")
		for _, t := range themes {
			fmt.Printf("  %s\n", t.Name)
		}
	}

	return nil
}
```

**Step 2: Create import command**

Create `tools/colorsync/cmd/importcmd.go`:

```go
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/importer"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

func init() {
	Register(Command{
		Name: "import",
		Help: "Import a theme (built-in name or .itermcolors file)",
		Run:  runImport,
	})
}

func runImport(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: colorsync import <name-or-file>")
	}

	source := args[0]
	var theme *palette.Theme
	var err error

	if strings.HasSuffix(source, ".itermcolors") {
		theme, err = importer.ParseItermColors(source)
	} else {
		theme, err = importer.GetBuiltin(source)
	}
	if err != nil {
		return err
	}

	dir := palette.ThemesDir()
	if err := palette.EnsureDir(dir); err != nil {
		return err
	}

	path := filepath.Join(dir, theme.Name+".json")
	if err := theme.Save(path); err != nil {
		return err
	}

	fmt.Printf("Imported %q -> %s\n", theme.Name, path)
	return nil
}
```

**Step 3: Create generate command**

Create `tools/colorsync/cmd/generate.go`:

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "generate",
		Help: "Generate a theme from bg, fg, and accent colors",
		Run:  runGenerate,
	})
}

func runGenerate(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	bg := prompt(reader, "Background (#hex): ")
	fg := prompt(reader, "Foreground (#hex): ")
	accent := prompt(reader, "Accent (#hex): ")
	name := prompt(reader, "Name: ")

	theme, err := palette.Generate(name, bg, fg, accent)
	if err != nil {
		return err
	}

	preview.Render(os.Stdout, theme)

	if confirm(reader, "Save? [y/n]: ") {
		dir := palette.ThemesDir()
		if err := palette.EnsureDir(dir); err != nil {
			return err
		}
		path := filepath.Join(dir, theme.Name+".json")
		if err := theme.Save(path); err != nil {
			return err
		}
		fmt.Printf("Saved to %s\n", path)
	}

	return nil
}

func prompt(r *bufio.Reader, msg string) string {
	fmt.Print(msg)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func confirm(r *bufio.Reader, msg string) bool {
	answer := prompt(r, msg)
	return strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes"
}
```

**Step 4: Create preview command**

Create `tools/colorsync/cmd/preview.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mhdev/dotfiles/tools/colorsync/importer"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "preview",
		Help: "Preview a theme's colors in the terminal",
		Run:  runPreview,
	})
}

func runPreview(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: colorsync preview <theme>")
	}

	theme, err := resolveTheme(args[0])
	if err != nil {
		return err
	}

	preview.Render(os.Stdout, theme)
	return nil
}

func resolveTheme(name string) (*palette.Theme, error) {
	// Try saved themes first
	saved := filepath.Join(palette.ThemesDir(), name+".json")
	if t, err := palette.Load(saved); err == nil {
		return t, nil
	}

	// Try built-in
	if t, err := importer.GetBuiltin(name); err == nil {
		return t, nil
	}

	return nil, fmt.Errorf("theme %q not found (use 'colorsync list' to see available themes)", name)
}
```

**Step 5: Create apply command**

Create `tools/colorsync/cmd/apply.go`:

```go
package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/exporter"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "apply",
		Help: "Apply a theme to neovim, tmux, and iTerm",
		Run:  runApply,
	})
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	targets := fs.String("target", "nvim,tmux,iterm", "Comma-separated targets: nvim,tmux,iterm")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		return fmt.Errorf("usage: colorsync apply <theme> [--target nvim,tmux,iterm]")
	}

	theme, err := resolveTheme(remaining[0])
	if err != nil {
		return err
	}

	preview.Render(os.Stdout, theme)

	reader := bufio.NewReader(os.Stdin)
	if !confirm(reader, "Apply this theme? [y/n]: ") {
		fmt.Println("Cancelled.")
		return nil
	}

	targetSet := make(map[string]bool)
	for _, t := range strings.Split(*targets, ",") {
		targetSet[strings.TrimSpace(t)] = true
	}

	if targetSet["nvim"] {
		path := exporter.NeovimDefaultPath(theme.Name)
		if err := exporter.ExportNeovim(theme, path); err != nil {
			return fmt.Errorf("neovim: %w", err)
		}
		fmt.Printf("Neovim: wrote %s\n", path)
		fmt.Printf("  Activate with: colorscheme %s\n", strings.ReplaceAll(theme.Name, "-", "_"))
	}

	if targetSet["tmux"] {
		path := exporter.TmuxDefaultPath()
		if err := exporter.ExportTmux(theme, path); err != nil {
			return fmt.Errorf("tmux: %w", err)
		}
		fmt.Printf("tmux: wrote %s\n", path)
		fmt.Println("  Add to .tmux.conf: source-file ~/.tmux/theme.conf")
		fmt.Println("  Reload with: tmux source-file ~/.tmux.conf")
	}

	if targetSet["iterm"] {
		filePath := exporter.ItermDefaultPath(theme.Name)
		if err := exporter.ExportItermFile(theme, filePath); err != nil {
			return fmt.Errorf("iterm file: %w", err)
		}
		fmt.Printf("iTerm: wrote %s\n", filePath)

		// Live-update running terminal
		exporter.WriteItermEscapes(os.Stdout, theme)
		fmt.Println("iTerm: live colors updated")
	}

	return nil
}
```

**Step 6: Update main.go to import cmd subpackages**

Update `tools/colorsync/main.go`:

```go
package main

import (
	"github.com/mhdev/dotfiles/tools/colorsync/cmd"

	_ "github.com/mhdev/dotfiles/tools/colorsync/cmd"
)

func main() {
	cmd.Execute()
}
```

Note: The `init()` functions in each cmd file auto-register. The blank import is not needed since `cmd` is already imported. The file stays as:

```go
package main

import "github.com/mhdev/dotfiles/tools/colorsync/cmd"

func main() {
	cmd.Execute()
}
```

**Step 7: Build and verify**

```bash
cd tools/colorsync && go build -o colorsync . && ./colorsync list
```

Expected: prints built-in themes list.

**Step 8: Commit**

```bash
git add tools/colorsync/cmd/ tools/colorsync/main.go
git commit -m "feat(colorsync): wire up all CLI commands (list, import, generate, preview, apply)"
```

---

### Task 11: CLAUDE.md

**Files:**
- Create: `tools/colorsync/CLAUDE.md`

**Step 1: Write CLAUDE.md**

Create `tools/colorsync/CLAUDE.md`:

```markdown
# colorsync

A Go CLI tool that syncs color schemes across neovim, tmux, and iTerm2.

## Quick Reference

```bash
# Build
cd tools/colorsync && go build -o colorsync .

# Run tests
cd tools/colorsync && go test ./... -v

# Commands
./colorsync list                          # List available themes
./colorsync import catppuccin-mocha       # Import a built-in theme
./colorsync import ~/Downloads/theme.itermcolors  # Import from file
./colorsync generate                      # Create theme from bg/fg/accent
./colorsync preview catppuccin-mocha      # Preview in terminal
./colorsync apply catppuccin-mocha        # Apply to all targets
./colorsync apply gruvbox-dark --target tmux,nvim  # Apply to specific targets
```

## Architecture

- **Palette model**: `palette/palette.go` - `Theme` struct with 18 colors (bg, fg, cursor, 16 ANSI). JSON serialization. Themes stored in `~/.config/colorsync/themes/`.
- **Color generation**: `palette/generate.go` - Derives 16 ANSI colors from bg/fg/accent using HSL manipulation.
- **Importers**: `importer/` - `builtin.go` has 6 hardcoded themes. `itermcolors.go` parses Apple plist XML.
- **Exporters**: `exporter/` - `neovim.go` writes standalone Lua colorscheme. `tmux.go` writes theme.conf. `iterm.go` writes .itermcolors and sends live escape sequences.
- **Preview**: `preview/preview.go` - Renders color swatches using 24-bit ANSI escapes.
- **CLI**: `cmd/` - Subcommands registered via `init()` functions. No cobra/viper, plain stdlib.

## Output Paths

| Target | Output | Activation |
|--------|--------|------------|
| neovim | `~/.config/nvim/colors/<name>.lua` | `:colorscheme <name>` |
| tmux | `~/.tmux/theme.conf` | `source-file ~/.tmux/theme.conf` in `.tmux.conf` |
| iTerm | `~/.config/colorsync/output/<name>.itermcolors` + live escape sequences | Automatic for current session |

## Conventions

- Go stdlib only, no external dependencies
- Theme names use hyphens (`catppuccin-mocha`), Lua filenames use underscores (`catppuccin_mocha.lua`)
- All colors stored as 7-char hex strings (`#rrggbb`)
- Tests live alongside source in `_test.go` files
- Built-in themes are hardcoded in `importer/builtin.go`

## Adding a New Built-in Theme

Edit `importer/builtin.go` and add an entry to the `builtins` map:

```go
"theme-name": {
    Name: "theme-name", Background: "#...", Foreground: "#...", Cursor: "#...",
    Colors: [16]string{ /* ansi 0-15 */ },
},
```

## Adding a New Export Target

1. Create `exporter/<target>.go` with an `Export<Target>(theme *palette.Theme, path string) error` function
2. Add a test in `exporter/<target>_test.go`
3. Wire into `cmd/apply.go` by adding a new target check in `runApply`
```

**Step 2: Commit**

```bash
git add tools/colorsync/CLAUDE.md
git commit -m "docs(colorsync): add CLAUDE.md for AI-assisted development"
```

---

### Task 12: Final Integration Test

**Step 1: Build the binary**

```bash
cd tools/colorsync && go build -o colorsync .
```

**Step 2: Run all tests**

```bash
cd tools/colorsync && go test ./... -v
```

Expected: all tests PASS.

**Step 3: Smoke test the CLI**

```bash
cd tools/colorsync && ./colorsync list && ./colorsync preview catppuccin-mocha
```

Expected: lists themes, then shows colored swatches for catppuccin-mocha.

**Step 4: Commit if any fixes were needed, then clean up binary**

```bash
cd tools/colorsync && rm -f colorsync
```
