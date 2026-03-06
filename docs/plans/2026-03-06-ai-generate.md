# AI Generate Command Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `ai-generate` command that uses Apple Foundation Models to generate terminal themes from natural language descriptions.

**Architecture:** A Swift CLI helper (`colorsync-ai`) uses Foundation Models guided generation (`@Generable`) to produce structured theme JSON. The Go CLI shells out to it, parses JSON, previews, and saves.

**Tech Stack:** Swift 6.2 + FoundationModels framework (macOS 26), Go stdlib

---

### Task 1: Create Swift Package Structure

**Files:**
- Create: `tools/colorsync-ai/Package.swift`
- Create: `tools/colorsync-ai/Sources/main.swift`

**Step 1: Create Package.swift**

```swift
// swift-tools-version: 6.2

import PackageDescription

let package = Package(
    name: "colorsync-ai",
    platforms: [.macOS(.v26)],
    targets: [
        .executableTarget(name: "colorsync-ai")
    ]
)
```

**Step 2: Create a minimal main.swift that compiles**

```swift
import Foundation
import FoundationModels

print("colorsync-ai placeholder")
```

**Step 3: Build to verify the package compiles**

Run: `cd tools/colorsync-ai && swift build 2>&1`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add tools/colorsync-ai/Package.swift tools/colorsync-ai/Sources/main.swift
git commit -m "feat(colorsync-ai): scaffold Swift package for AI theme generation"
```

---

### Task 2: Implement Swift Generable Theme Struct and Generation

**Files:**
- Modify: `tools/colorsync-ai/Sources/main.swift`

**Step 1: Implement the full main.swift**

Replace `main.swift` with the complete implementation:

```swift
import Foundation
import FoundationModels

@Generable(description: "A terminal color theme with background, foreground, cursor, and 16 ANSI colors")
struct GeneratedTheme {
    @Guide(description: "A short hyphenated name for the theme, e.g. 'autumn-dusk'")
    var name: String

    @Guide(description: "Background color as a 7-character hex string like #1a1b26. For dark themes use a dark color, for light themes use a light color.")
    var background: String

    @Guide(description: "Foreground/text color as a 7-character hex string like #cdd6f4. Should contrast well with the background.")
    var foreground: String

    @Guide(description: "Cursor color as a 7-character hex string. Usually same as foreground or an accent color.")
    var cursor: String

    @Guide(description: """
        Exactly 16 ANSI terminal colors as 7-character hex strings (#rrggbb). \
        Index meanings: 0=black, 1=red, 2=green, 3=yellow, 4=blue, 5=magenta, 6=cyan, 7=white, \
        8=bright black, 9=bright red, 10=bright green, 11=bright yellow, 12=bright blue, \
        13=bright magenta, 14=bright cyan, 15=bright white. \
        Colors should be harmonious and match the theme description.
        """)
    @Guide(.count(16))
    var colors: [String]
}

@main
struct ColorsyncAI {
    static func main() async throws {
        let args = CommandLine.arguments
        guard args.count >= 2 else {
            FileHandle.standardError.write(Data("Usage: colorsync-ai <description>\n".utf8))
            exit(1)
        }

        let description = args.dropFirst().joined(separator: " ")

        let model = SystemLanguageModel.default
        guard model.isAvailable else {
            FileHandle.standardError.write(Data("Error: Apple Intelligence is not available on this device.\n".utf8))
            exit(2)
        }

        let session = LanguageModelSession()
        let prompt = """
            Generate a terminal color theme based on this description: \(description)

            The theme needs a name, background color, foreground color, cursor color, \
            and exactly 16 ANSI colors. All colors must be 7-character hex strings starting with #. \
            The colors should be harmonious, visually appealing, and match the description. \
            Ensure good contrast between background and foreground.
            """

        let response = try await session.respond(
            to: prompt,
            generating: GeneratedTheme.self
        )

        let theme = response.content

        // Build JSON matching Go's palette.Theme format
        var dict: [String: Any] = [
            "name": theme.name,
            "background": theme.background,
            "foreground": theme.foreground,
            "cursor": theme.cursor,
            "colors": theme.colors
        ]

        let jsonData = try JSONSerialization.data(withJSONObject: dict, options: [.prettyPrinted, .sortedKeys])
        FileHandle.standardOutput.write(jsonData)
    }
}
```

**Step 2: Build and verify**

Run: `cd tools/colorsync-ai && swift build 2>&1`
Expected: Build succeeds

**Step 3: Test the binary manually**

Run: `cd tools/colorsync-ai && .build/debug/colorsync-ai "a warm dark theme inspired by autumn forests"`
Expected: JSON output with name, background, foreground, cursor, and 16 colors

**Step 4: Commit**

```bash
git add tools/colorsync-ai/Sources/main.swift
git commit -m "feat(colorsync-ai): implement Foundation Models guided theme generation"
```

---

### Task 3: Add Go ai-generate Command

**Files:**
- Create: `tools/colorsync/cmd/ai_generate.go`

**Step 1: Implement ai_generate.go**

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "ai-generate",
		Help: "Generate a theme from a natural language description using Apple Intelligence",
		Run:  runAIGenerate,
	})
}

func runAIGenerate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: colorsync ai-generate <description>\n  example: colorsync ai-generate \"a warm dark theme inspired by autumn\"")
	}

	description := strings.Join(args, " ")
	fmt.Printf("Generating theme: %q\n", description)

	bin, err := findAIBinary()
	if err != nil {
		return err
	}

	cmd := exec.Command(bin, description)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("colorsync-ai failed: %w", err)
	}

	var theme palette.Theme
	if err := json.Unmarshal(out, &theme); err != nil {
		return fmt.Errorf("parsing AI output: %w", err)
	}

	preview.Render(os.Stdout, &theme)

	reader := bufioReader()
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

func findAIBinary() (string, error) {
	// Look adjacent to this binary first
	self, err := os.Executable()
	if err == nil {
		adjacent := filepath.Join(filepath.Dir(self), "colorsync-ai")
		if _, err := os.Stat(adjacent); err == nil {
			return adjacent, nil
		}
	}

	// Fall back to PATH
	p, err := exec.LookPath("colorsync-ai")
	if err != nil {
		return "", fmt.Errorf("colorsync-ai not found; build it with: cd tools/colorsync-ai && swift build")
	}
	return p, nil
}
```

Note: The `confirm` function already exists in `cmd/generate.go`. We need a helper to create a `bufio.Reader` from stdin — but `confirm` and `prompt` already exist there. We just need to add the `bufioReader` helper or use `bufio.NewReader(os.Stdin)` inline.

Actually, looking at `generate.go`, `confirm` takes a `*bufio.Reader`. So `ai_generate.go` should just do:

```go
reader := bufio.NewReader(os.Stdin)
if confirm(reader, "Save? [y/n]: ") {
```

And add `"bufio"` to imports. Remove the `bufioReader()` call.

**Step 2: Build and verify**

Run: `cd tools/colorsync && go build -o colorsync . 2>&1`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add tools/colorsync/cmd/ai_generate.go
git commit -m "feat(colorsync): add ai-generate command using Apple Intelligence"
```

---

### Task 4: End-to-End Test

**Step 1: Build both binaries**

```bash
cd tools/colorsync-ai && swift build
cd tools/colorsync && go build -o colorsync .
```

**Step 2: Run ai-generate**

```bash
cd tools/colorsync
./colorsync ai-generate "a cozy dark theme inspired by a warm autumn forest"
```

Expected: Preview renders in terminal, prompt asks to save, JSON saved to `~/.config/colorsync/themes/<name>.json`

**Step 3: Verify the saved theme works with apply**

```bash
./colorsync preview <generated-name>
./colorsync apply <generated-name> --target nvim
```

**Step 4: Commit any fixes**

---

### Task 5: Update CLAUDE.md

**Files:**
- Modify: `tools/colorsync/CLAUDE.md`

**Step 1: Add ai-generate to command table and docs**

Add to the Quick Reference section:
```
./colorsync ai-generate "warm autumn theme"  # AI-generate from description
```

Add to Architecture section:
```
- **AI helper**: `../colorsync-ai/` - Swift CLI using Apple Foundation Models for AI theme generation.
```

**Step 2: Commit**

```bash
git add tools/colorsync/CLAUDE.md
git commit -m "docs(colorsync): add ai-generate to CLAUDE.md"
```
