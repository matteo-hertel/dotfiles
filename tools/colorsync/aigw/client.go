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

// DefaultBaseURL is the Vercel AI Gateway endpoint.
const DefaultBaseURL = "https://ai-gateway.vercel.sh/v1"

// Model represents an AI model available through the gateway.
type Model struct {
	ID   string
	Name string
}

// Models returns the list of available AI models.
func Models() []Model {
	return []Model{
		{ID: "anthropic/claude-opus-4-6", Name: "Claude Opus 4.6"},
		{ID: "openai/gpt-5.4", Name: "GPT-5.4"},
		{ID: "google/gemini-3.1-pro", Name: "Gemini 3.1 Pro"},
	}
}

// themeSchema is the JSON schema that constrains the model output to match palette.Theme.
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
	"required":             []string{"name", "background", "foreground", "cursor", "colors"},
	"additionalProperties": false,
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

// ProgressFunc is called during streaming with the cumulative token count.
type ProgressFunc func(tokens int)

// Generate calls the AI gateway to generate a terminal color theme from a description.
// It streams the response using SSE and calls the progress callback with cumulative token counts.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
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
		return nil, fmt.Errorf("connecting to AI gateway at %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI gateway returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse SSE stream
	var content strings.Builder
	tokens := 0
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		// Check context between lines
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request timed out")
		}

		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Only process data lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			break
		}

		// Parse the SSE chunk (OpenAI format)
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // skip malformed chunks
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
		return nil, fmt.Errorf("reading response stream: %w", err)
	}

	result := content.String()
	if result == "" {
		return nil, fmt.Errorf("empty response from model — try a different model or description")
	}

	var theme palette.Theme
	if err := json.Unmarshal([]byte(result), &theme); err != nil {
		return nil, fmt.Errorf("parsing theme from model output: %w (raw: %s)", err, result)
	}

	return &theme, nil
}
