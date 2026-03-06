# AI Generate with Ollama Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite `ai-generate` to use Ollama (local LLM) instead of Apple Foundation Models, delete the Swift helper, add setup instructions and .gitignore.

**Architecture:** The Go CLI calls Ollama's REST API (`POST localhost:11434/api/chat`) with a JSON schema matching `palette.Theme`. The model generates the full 19-color theme directly. No external binaries, no new dependencies — just `net/http` + `encoding/json`.

**Tech Stack:** Go stdlib (`net/http`, `encoding/json`, `flag`), Ollama REST API

---

### Task 1: Delete Swift package and clean up references

**Files:**
- Delete: `tools/colorsync-ai/` (entire directory)
- Modify: `tools/colorsync/CLAUDE.md`

**Step 1: Delete the Swift package directory**

```bash
rm -rf tools/colorsync-ai
```

**Step 2: Remove the colorsync-ai binary from ~/go/bin if present**

```bash
rm -f ~/go/bin/colorsync-ai
```

**Step 3: Verify it's gone**

Run: `ls tools/colorsync-ai 2>&1`
Expected: `No such file or directory`

**Step 4: Commit**

```bash
git add -A tools/colorsync-ai
git commit -m "chore: remove Swift colorsync-ai package (replaced by Ollama)"
```

---

### Task 2: Add .gitignore for build artifacts

**Files:**
- Create: `tools/colorsync/.gitignore`

**Step 1: Create .gitignore**

```gitignore
# Compiled binary
colorsync
```

**Step 2: Remove tracked binary if present**

```bash
cd tools/colorsync
git rm --cached colorsync 2>/dev/null || true
```

**Step 3: Commit**

```bash
git add tools/colorsync/.gitignore
git commit -m "chore: add .gitignore for colorsync build artifacts"
```

---

### Task 3: Create Ollama client package

**Files:**
- Create: `tools/colorsync/ollama/client.go`
- Create: `tools/colorsync/ollama/client_test.go`

This package encapsulates all Ollama HTTP interaction so `cmd/ai_generate.go` stays clean.

**Step 1: Write the test**

```go
package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerate(t *testing.T) {
	// Mock Ollama API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/chat" {
			t.Fatalf("expected /api/chat, got %s", r.URL.Path)
		}

		// Decode request to verify structure
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		model, _ := req["model"].(string)
		if model == "" {
			t.Fatal("expected model in request")
		}
		if req["stream"] != false {
			t.Fatal("expected stream: false")
		}
		if req["format"] == nil {
			t.Fatal("expected format (JSON schema) in request")
		}

		// Return a valid theme response
		resp := map[string]any{
			"message": map[string]any{
				"role": "assistant",
				"content": `{"name":"test-theme","background":"#1a1b26","foreground":"#c0caf5","cursor":"#c0caf5","colors":["#15161e","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#a9b1d6","#414868","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#c0caf5"]}`,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	theme, err := Generate(server.URL, "qwen3:32b", "a dark blue theme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if theme.Name != "test-theme" {
		t.Errorf("expected name 'test-theme', got %q", theme.Name)
	}
	if theme.Background != "#1a1b26" {
		t.Errorf("expected background '#1a1b26', got %q", theme.Background)
	}
	if len(theme.Colors) != 16 {
		t.Errorf("expected 16 colors, got %d", len(theme.Colors))
	}
}

func TestGenerateConnectionRefused(t *testing.T) {
	_, err := Generate("http://localhost:1", "qwen3:32b", "test")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestGenerateInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"message": map[string]any{
				"role":    "assistant",
				"content": `not valid json`,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	_, err := Generate(server.URL, "qwen3:32b", "test")
	if err == nil {
		t.Fatal("expected error for invalid JSON content")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd tools/colorsync && go test ./ollama/ -v`
Expected: FAIL (package doesn't exist yet)

**Step 3: Write the implementation**

```go
package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

const DefaultURL = "http://localhost:11434"
const DefaultModel = "qwen3:32b"

// themeSchema is the JSON schema that constrains Ollama's output to match palette.Theme.
var themeSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"name":       map[string]any{"type": "string"},
		"background": map[string]any{"type": "string"},
		"foreground": map[string]any{"type": "string"},
		"cursor":     map[string]any{"type": "string"},
		"colors": map[string]any{
			"type":     "array",
			"items":    map[string]any{"type": "string"},
			"minItems": 16,
			"maxItems": 16,
		},
	},
	"required": []string{"name", "background", "foreground", "cursor", "colors"},
}

const systemPrompt = `You are a terminal color theme designer. You generate complete terminal color themes as JSON.

Rules:
- All colors MUST be 7-character hex strings starting with # (e.g. #1a1b26)
- "name" should be a short, hyphenated, lowercase name matching the theme mood (e.g. "autumn-dusk")
- "background" is the terminal background color
- "foreground" is the default text color — must contrast well with background
- "cursor" is the cursor color — typically the foreground or an accent color
- "colors" is exactly 16 ANSI terminal colors as hex strings:
  [0]=black [1]=red [2]=green [3]=yellow [4]=blue [5]=magenta [6]=cyan [7]=white
  [8]=bright black [9]=bright red [10]=bright green [11]=bright yellow
  [12]=bright blue [13]=bright magenta [14]=bright cyan [15]=bright white
- Colors should be harmonious, visually appealing, and match the description
- For dark themes: background should be dark, foreground light
- For light themes: background should be light, foreground dark
- Bright variants (8-15) should be lighter/more saturated versions of normal (0-7)`

// Generate calls Ollama to generate a terminal color theme from a description.
func Generate(baseURL, model, description string) (*palette.Theme, error) {
	prompt := fmt.Sprintf("Generate a terminal color theme based on this description: %s", description)

	reqBody := map[string]any{
		"model":  model,
		"stream": false,
		"format": themeSchema,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("connecting to Ollama at %s: %w\n\nIs Ollama running? Start it with: ollama serve", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decoding Ollama response: %w", err)
	}

	var theme palette.Theme
	if err := json.Unmarshal([]byte(chatResp.Message.Content), &theme); err != nil {
		return nil, fmt.Errorf("parsing theme from model output: %w", err)
	}

	return &theme, nil
}
```

**Step 4: Run tests**

Run: `cd tools/colorsync && go test ./ollama/ -v`
Expected: All 3 tests PASS

**Step 5: Commit**

```bash
git add tools/colorsync/ollama/
git commit -m "feat(colorsync): add Ollama client for AI theme generation"
```

---

### Task 4: Rewrite ai-generate command

**Files:**
- Rewrite: `tools/colorsync/cmd/ai_generate.go`

**Step 1: Rewrite ai_generate.go**

Replace the entire file:

```go
package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/ollama"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "ai-generate",
		Help: "Generate a theme from a natural language description using Ollama",
		Run:  runAIGenerate,
	})
}

func runAIGenerate(args []string) error {
	fs := flag.NewFlagSet("ai-generate", flag.ExitOnError)
	model := fs.String("model", ollama.DefaultModel, "Ollama model to use")
	url := fs.String("url", ollama.DefaultURL, "Ollama API URL")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("usage: colorsync ai-generate [--model qwen3:32b] [--url http://localhost:11434] <description>\n  example: colorsync ai-generate \"a warm dark theme inspired by autumn\"")
	}

	description := strings.Join(remaining, " ")
	fmt.Printf("Generating theme with %s: %q\n", *model, description)
	fmt.Println("This may take a moment...")

	theme, err := ollama.Generate(*url, *model, description)
	if err != nil {
		return err
	}

	if theme.Name == "" {
		return fmt.Errorf("model generated a theme with an empty name")
	}

	preview.Render(os.Stdout, theme)

	reader := bufio.NewReader(os.Stdin)
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
```

**Step 2: Build**

Run: `cd tools/colorsync && go build -o colorsync .`
Expected: Build succeeds

**Step 3: Run existing tests**

Run: `cd tools/colorsync && go test ./... -v`
Expected: All tests pass (including new ollama tests)

**Step 4: Commit**

```bash
git add tools/colorsync/cmd/ai_generate.go
git commit -m "feat(colorsync): rewrite ai-generate to use Ollama instead of Apple FM"
```

---

### Task 5: Update CLAUDE.md with setup instructions and remove Swift references

**Files:**
- Rewrite: `tools/colorsync/CLAUDE.md`

**Step 1: Update CLAUDE.md**

Remove all Swift/colorsync-ai references. Add Ollama setup instructions. The updated file should have:

- Remove `cd tools/colorsync-ai && swift build` from Build section
- Change `ai-generate` description from "(macOS 26+)" to "(requires Ollama)"
- Remove the **AI helper** line from Architecture
- Add `--model` and `--url` flags to ai-generate example
- Add a new "## AI Generate Setup" section with Ollama install and model pull instructions:

```markdown
## AI Generate Setup

The `ai-generate` command uses [Ollama](https://ollama.com) to run a local LLM.

### One-time setup

1. Install Ollama:
   ```bash
   brew install ollama
   ```

2. Start the Ollama server (or use the Ollama app):
   ```bash
   ollama serve
   ```

3. Pull the default model (~20GB download):
   ```bash
   ollama pull qwen3:32b
   ```

### Usage

```bash
colorsync ai-generate "a warm dark theme inspired by autumn forests"
colorsync ai-generate --model gemma3:27b "cool blue cyberpunk theme"
colorsync ai-generate --model llama3.1:8b "minimal grayscale theme"
```

Any Ollama model that supports structured output works. Larger models produce better color palettes.
```

**Step 2: Commit**

```bash
git add tools/colorsync/CLAUDE.md
git commit -m "docs(colorsync): update CLAUDE.md with Ollama setup, remove Swift refs"
```

---

### Task 6: End-to-end test

**Step 1: Build**

```bash
cd tools/colorsync && go build -o colorsync .
```

**Step 2: Verify Ollama is running**

```bash
curl -s http://localhost:11434/api/tags | head -5
```

Expected: JSON response listing available models

**Step 3: Run ai-generate**

```bash
echo "n" | ./colorsync ai-generate "a cozy dark theme inspired by autumn forests"
```

Expected: Theme preview renders, prompt asks to save

**Step 4: Test with --model flag**

```bash
echo "n" | ./colorsync ai-generate --model qwen3:32b "cool ocean blue theme"
```

Expected: Same flow, different theme

**Step 5: Test error case (Ollama not running)**

Stop Ollama and run:
```bash
./colorsync ai-generate "test" 2>&1
```

Expected: Clear error message about Ollama not being reachable

**Step 6: Install**

```bash
cd tools/colorsync && go install .
```

**Step 7: Commit any fixes**
