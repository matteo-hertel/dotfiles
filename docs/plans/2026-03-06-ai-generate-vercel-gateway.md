# AI Generate via Vercel AI Gateway — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace Ollama-based `ai-generate` with Vercel AI Gateway, running 3 frontier models in parallel and letting the user pick their favorite result.

**Architecture:** New `aigw/` package wraps the OpenAI-compatible Vercel AI Gateway REST API (`POST /v1/chat/completions`). The `cmd/ai_generate.go` command fires 3 concurrent requests (Claude Opus 4.6, GPT-5.4, Gemini 3.1 Pro), shows live per-model status, waits for all to finish, renders previews, and prompts the user to pick one.

**Tech Stack:** Go stdlib only (`net/http`, `encoding/json`, `bufio`, `context`, `sync`). Vercel AI Gateway with `AI_GATEWAY_API_KEY` env var.

---

### Task 1: Delete the `ollama/` package

**Files:**
- Delete: `tools/colorsync/ollama/client.go`
- Delete: `tools/colorsync/ollama/client_test.go`

**Step 1: Remove the files**

```bash
cd tools/colorsync && rm -rf ollama/
```

**Step 2: Verify build still works (it won't — ai_generate.go imports ollama)**

```bash
cd tools/colorsync && go build ./... 2>&1 | head -5
```

Expected: compile error referencing `ollama` import in `cmd/ai_generate.go`. That's fine — we fix it in Task 3.

**Step 3: Commit**

```bash
git add -A tools/colorsync/ollama/ && git commit -m "chore(colorsync): remove ollama package"
```

---

### Task 2: Create `aigw/client.go` with tests

**Files:**
- Create: `tools/colorsync/aigw/client.go`
- Create: `tools/colorsync/aigw/client_test.go`

**Step 1: Write the test file**

`tools/colorsync/aigw/client_test.go`:

```go
package aigw

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testThemeJSON = `{"name":"test-theme","background":"#1a1b26","foreground":"#c0caf5","cursor":"#c0caf5","colors":["#15161e","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#a9b1d6","#414868","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#c0caf5"]}`

// sseChunk formats a Server-Sent Events data line (OpenAI streaming format).
func sseChunk(content string, done bool) string {
	if done {
		return "data: [DONE]\n\n"
	}
	chunk := map[string]any{
		"choices": []map[string]any{
			{"delta": map[string]string{"content": content}},
		},
	}
	b, _ := json.Marshal(chunk)
	return "data: " + string(b) + "\n\n"
}

func TestGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("expected /v1/chat/completions, got %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Fatalf("expected Bearer test-key, got %q", auth)
		}

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		if req["stream"] != true {
			t.Fatal("expected stream: true")
		}
		if req["response_format"] == nil {
			t.Fatal("expected response_format in request")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		// Stream in a few chunks
		fmt.Fprint(w, sseChunk(`{"name":"test-theme","background":"#1a1b26",`, false))
		flusher.Flush()
		fmt.Fprint(w, sseChunk(`"foreground":"#c0caf5","cursor":"#c0caf5",`, false))
		flusher.Flush()
		fmt.Fprint(w, sseChunk(`"colors":["#15161e","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#a9b1d6","#414868","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#c0caf5"]}`, false))
		flusher.Flush()
		fmt.Fprint(w, sseChunk("", true))
		flusher.Flush()
	}))
	defer server.Close()

	var tokens int
	theme, err := Generate(context.Background(), server.URL+"/v1", "test-key", "openai/gpt-5.4", "dark blue theme", func(n int) {
		tokens = n
	})
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
	if tokens == 0 {
		t.Error("expected progress callback to be called")
	}
}

func TestGenerateConnectionRefused(t *testing.T) {
	_, err := Generate(context.Background(), "http://localhost:1/v1", "key", "openai/gpt-5.4", "test", nil)
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestGenerateTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := Generate(ctx, server.URL+"/v1", "key", "openai/gpt-5.4", "test", nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestGenerateInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseChunk("not valid json at all", false))
		fmt.Fprint(w, sseChunk("", true))
	}))
	defer server.Close()

	_, err := Generate(context.Background(), server.URL+"/v1", "key", "openai/gpt-5.4", "test", nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON content")
	}
}

func TestGenerateAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"message":"invalid api key"}}`)
	}))
	defer server.Close()

	_, err := Generate(context.Background(), server.URL+"/v1", "bad-key", "openai/gpt-5.4", "test", nil)
	if err == nil {
		t.Fatal("expected error for auth failure")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got: %v", err)
	}
}

func TestModels(t *testing.T) {
	models := Models()
	if len(models) != 3 {
		t.Fatalf("expected 3 models, got %d", len(models))
	}
	// Verify all have required fields
	for _, m := range models {
		if m.ID == "" || m.Name == "" {
			t.Errorf("model missing ID or Name: %+v", m)
		}
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd tools/colorsync && go test ./aigw/ -v 2>&1 | head -5
```

Expected: FAIL — package doesn't exist yet.

**Step 3: Write the implementation**

`tools/colorsync/aigw/client.go`:

```go
package aigw

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

const DefaultBaseURL = "https://ai-gateway.vercel.sh/v1"

// Model represents an AI model available through the gateway.
type Model struct {
	ID   string // e.g. "anthropic/claude-opus-4-6"
	Name string // e.g. "Claude Opus 4.6"
}

// Models returns the three frontier models used for parallel generation.
func Models() []Model {
	return []Model{
		{ID: "anthropic/claude-opus-4-6", Name: "Claude Opus 4.6"},
		{ID: "openai/gpt-5.4", Name: "GPT-5.4"},
		{ID: "google/gemini-3.1-pro", Name: "Gemini 3.1 Pro"},
	}
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

// themeSchema is the JSON schema for structured output.
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

// ProgressFunc is called with the number of streaming chunks received so far.
type ProgressFunc func(tokens int)

// Generate calls the Vercel AI Gateway to generate a theme.
// It streams the response and calls progress as chunks arrive.
func Generate(ctx context.Context, baseURL, apiKey, model, description string, progress ProgressFunc) (*palette.Theme, error) {
	prompt := fmt.Sprintf("Generate a terminal color theme based on this description: %s", description)

	reqBody := map[string]any{
		"model":  model,
		"stream": true,
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "theme",
				"strict": true,
				"schema": themeSchema,
			},
		},
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request timed out")
		}
		return nil, fmt.Errorf("connecting to AI Gateway: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI Gateway returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse SSE stream
	var content strings.Builder
	tokens := 0
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			content.WriteString(chunk.Choices[0].Delta.Content)
			tokens++
			if progress != nil {
				progress(tokens)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request timed out")
		}
		return nil, fmt.Errorf("reading stream: %w", err)
	}

	if content.Len() == 0 {
		return nil, fmt.Errorf("model returned empty response")
	}

	var theme palette.Theme
	if err := json.Unmarshal([]byte(content.String()), &theme); err != nil {
		return nil, fmt.Errorf("parsing theme JSON: %w\nraw output: %s", err, content.String())
	}

	return &theme, nil
}
```

**Step 4: Run tests**

```bash
cd tools/colorsync && go test ./aigw/ -v
```

Expected: all 6 tests PASS.

**Step 5: Commit**

```bash
git add tools/colorsync/aigw/ && git commit -m "feat(colorsync): add aigw package for Vercel AI Gateway"
```

---

### Task 3: Rewrite `cmd/ai_generate.go` with parallel execution and live status

**Files:**
- Modify: `tools/colorsync/cmd/ai_generate.go` (complete rewrite)

**Step 1: Write the new command**

Replace the entire contents of `tools/colorsync/cmd/ai_generate.go`:

```go
package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mhdev/dotfiles/tools/colorsync/aigw"
	"github.com/mhdev/dotfiles/tools/colorsync/palette"
	"github.com/mhdev/dotfiles/tools/colorsync/preview"
)

func init() {
	Register(Command{
		Name: "ai-generate",
		Help: "Generate a theme from a description using 3 AI models in parallel",
		Run:  runAIGenerate,
	})
}

// modelStatus tracks the live state of a single model's generation.
type modelStatus struct {
	model  aigw.Model
	state  string // "waiting", "streaming", "done", "failed"
	tokens int
	err    error
	theme  *palette.Theme
	dur    time.Duration
}

func runAIGenerate(args []string) error {
	fs := flag.NewFlagSet("ai-generate", flag.ExitOnError)
	baseURL := fs.String("url", aigw.DefaultBaseURL, "AI Gateway base URL")
	timeout := fs.Duration("timeout", 2*time.Minute, "Per-model timeout")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("usage: colorsync ai-generate [--url URL] [--timeout 2m] <description>\n  example: colorsync ai-generate \"a warm dark theme inspired by autumn\"")
	}

	apiKey := os.Getenv("AI_GATEWAY_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("AI_GATEWAY_API_KEY environment variable is not set\n\n  Get your key at: https://vercel.com/docs/ai-gateway\n  Then: export AI_GATEWAY_API_KEY=your-key-here")
	}

	description := strings.Join(remaining, " ")
	models := aigw.Models()

	fmt.Printf("Generating theme: %q\n", description)
	fmt.Printf("Racing %d models...\n\n", len(models))

	// Initialize status for each model
	statuses := make([]modelStatus, len(models))
	var mu sync.Mutex
	for i, m := range models {
		statuses[i] = modelStatus{model: m, state: "waiting"}
	}

	// Spinner characters
	spinChars := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	done := make(chan struct{})
	start := time.Now()

	// Display goroutine — redraws all model statuses
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				mu.Lock()
				var lines []string
				for _, s := range statuses {
					lines = append(lines, formatStatus(s, spinChars[i%len(spinChars)]))
				}
				mu.Unlock()

				// Move cursor up and clear, then redraw
				if i > 0 {
					fmt.Printf("\033[%dA", len(lines))
				}
				for _, line := range lines {
					fmt.Printf("\033[K%s\n", line)
				}

				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Launch parallel generation
	var wg sync.WaitGroup
	for i, m := range models {
		wg.Add(1)
		go func(idx int, model aigw.Model) {
			defer wg.Done()
			modelStart := time.Now()

			mu.Lock()
			statuses[idx].state = "streaming"
			mu.Unlock()

			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			defer cancel()

			theme, err := aigw.Generate(ctx, *baseURL, apiKey, model.ID, description, func(tokens int) {
				mu.Lock()
				statuses[idx].tokens = tokens
				mu.Unlock()
			})

			mu.Lock()
			statuses[idx].dur = time.Since(modelStart).Truncate(time.Second)
			if err != nil {
				statuses[idx].state = "failed"
				statuses[idx].err = err
			} else {
				statuses[idx].state = "done"
				statuses[idx].theme = theme
			}
			mu.Unlock()
		}(i, m)
	}

	wg.Wait()
	close(done)
	time.Sleep(150 * time.Millisecond) // let final render flush

	// Final status redraw (static)
	fmt.Printf("\033[%dA", len(statuses))
	for _, s := range statuses {
		fmt.Printf("\033[K%s\n", formatStatus(s, ' '))
	}
	fmt.Println()

	elapsed := time.Since(start).Truncate(time.Second)
	fmt.Printf("All models finished in %s\n\n", elapsed)

	// Collect successful results
	var results []modelStatus
	for _, s := range statuses {
		if s.state == "done" && s.theme != nil {
			results = append(results, s)
		}
	}

	if len(results) == 0 {
		fmt.Println("All models failed:")
		for _, s := range statuses {
			if s.err != nil {
				fmt.Printf("  %s: %v\n", s.model.Name, s.err)
			}
		}
		return fmt.Errorf("no themes generated")
	}

	// Show numbered previews
	for i, r := range results {
		fmt.Printf("--- [%d] %s (%s) ---", i+1, r.model.Name, r.dur)
		preview.Render(os.Stdout, r.theme)
	}

	// Prompt user to pick
	reader := bufio.NewReader(os.Stdin)
	var chosen *palette.Theme
	if len(results) == 1 {
		chosen = results[0].theme
		fmt.Printf("Only one result from %s.\n", results[0].model.Name)
		if !confirm(reader, "Save? [y/n]: ") {
			return nil
		}
	} else {
		fmt.Printf("Pick a theme [1-%d] or 0 to discard: ", len(results))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		pick := 0
		fmt.Sscanf(input, "%d", &pick)
		if pick < 1 || pick > len(results) {
			fmt.Println("Discarded.")
			return nil
		}
		chosen = results[pick-1].theme
	}

	if chosen.Name == "" {
		return fmt.Errorf("model generated a theme with an empty name")
	}

	dir := palette.ThemesDir()
	if err := palette.EnsureDir(dir); err != nil {
		return err
	}
	path := filepath.Join(dir, chosen.Name+".json")
	if err := chosen.Save(path); err != nil {
		return err
	}
	fmt.Printf("Saved to %s\n", path)

	return nil
}

func formatStatus(s modelStatus, spin rune) string {
	switch s.state {
	case "waiting":
		return fmt.Sprintf("  %c  %-20s waiting...", spin, s.model.Name)
	case "streaming":
		if s.tokens > 0 {
			return fmt.Sprintf("  %c  %-20s streaming... %d tokens", spin, s.model.Name, s.tokens)
		}
		return fmt.Sprintf("  %c  %-20s connecting...", spin, s.model.Name)
	case "done":
		return fmt.Sprintf("  ✓  %-20s done (%s, %d tokens)", s.model.Name, s.dur, s.tokens)
	case "failed":
		msg := s.err.Error()
		if len(msg) > 60 {
			msg = msg[:60] + "..."
		}
		return fmt.Sprintf("  ✗  %-20s failed: %s", s.model.Name, msg)
	default:
		return fmt.Sprintf("  ?  %-20s %s", s.model.Name, s.state)
	}
}
```

**Step 2: Verify it builds**

```bash
cd tools/colorsync && go build ./...
```

Expected: clean build, no errors.

**Step 3: Run all tests**

```bash
cd tools/colorsync && go test ./... -v -timeout 30s
```

Expected: all packages pass.

**Step 4: Commit**

```bash
git add tools/colorsync/cmd/ai_generate.go && git commit -m "feat(colorsync): rewrite ai-generate for Vercel AI Gateway with parallel models"
```

---

### Task 4: Update CLAUDE.md documentation

**Files:**
- Modify: `tools/colorsync/CLAUDE.md`

**Step 1: Update the docs**

Replace the `## Architecture` bullet about Ollama:

```
- **Ollama client**: `ollama/client.go` - ...
```

With:

```
- **AI Gateway client**: `aigw/client.go` - Calls Vercel AI Gateway (OpenAI-compatible REST API) with structured JSON output. Streams responses. Used by `cmd/ai_generate.go`.
```

Replace the entire `## AI Generate Setup` section with:

```markdown
## AI Generate Setup

The `ai-generate` command uses the [Vercel AI Gateway](https://vercel.com/ai-gateway) to run 3 frontier models in parallel (Claude Opus 4.6, GPT-5.4, Gemini 3.1 Pro) and lets you pick the best result.

### One-time setup

1. Get an API key at [vercel.com/ai-gateway](https://vercel.com/docs/ai-gateway)

2. Export it:
   ```bash
   export AI_GATEWAY_API_KEY=your-key-here
   ```

### Usage

```bash
colorsync ai-generate "a warm dark theme inspired by autumn forests"
colorsync ai-generate "cool blue cyberpunk neon"
colorsync ai-generate "minimal grayscale with a hint of green"
```

### Flags

- `--url <url>` — AI Gateway base URL (default: `https://ai-gateway.vercel.sh/v1`)
- `--timeout <duration>` — Per-model timeout (default: `2m`)
```

Update the Quick Reference section to replace Ollama examples:

```bash
./colorsync ai-generate "warm autumn"       # AI-generate with 3 models in parallel
```

**Step 2: Verify nothing references ollama**

```bash
cd tools/colorsync && grep -ri ollama . --include="*.go" --include="*.md"
```

Expected: no results.

**Step 3: Commit**

```bash
git add tools/colorsync/CLAUDE.md && git commit -m "docs(colorsync): update CLAUDE.md for Vercel AI Gateway"
```

---

### Task 5: Build, install, and smoke test

**Step 1: Build**

```bash
cd tools/colorsync && go build -o colorsync . && go install .
```

**Step 2: Run without API key to verify error message**

```bash
unset AI_GATEWAY_API_KEY && ./colorsync ai-generate "test"
```

Expected: clear error about missing `AI_GATEWAY_API_KEY`.

**Step 3: Run with API key (manual smoke test)**

```bash
export AI_GATEWAY_API_KEY=your-key && ./colorsync ai-generate "warm autumn forest"
```

Expected output shape:
```
Generating theme: "warm autumn forest"
Racing 3 models...

  ⠹  Claude Opus 4.6      streaming... 12 tokens
  ⠼  GPT-5.4              streaming... 8 tokens
  ⠧  Gemini 3.1 Pro       streaming... 15 tokens

  ✓  Claude Opus 4.6      done (8s, 47 tokens)
  ✓  GPT-5.4              done (5s, 34 tokens)
  ✓  Gemini 3.1 Pro       done (6s, 41 tokens)

All models finished in 8s

--- [1] Claude Opus 4.6 (8s) ---
  Theme: autumn-ember
  ...color preview...

--- [2] GPT-5.4 (5s) ---
  Theme: forest-dusk
  ...color preview...

--- [3] Gemini 3.1 Pro (6s) ---
  Theme: warm-canopy
  ...color preview...

Pick a theme [1-3] or 0 to discard:
```

**Step 4: Final commit**

```bash
git add -A && git commit -m "chore(colorsync): build verification"
```
