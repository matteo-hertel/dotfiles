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
