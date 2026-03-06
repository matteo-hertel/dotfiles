package aigw

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// sseChunk formats content as an SSE data line. If done is true, it returns the stream terminator.
func sseChunk(content string, done bool) string {
	if done {
		return "data: [DONE]\n\n"
	}
	return fmt.Sprintf("data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", content)
}

func TestGenerate(t *testing.T) {
	// Build a valid theme JSON and split it into SSE chunks
	themeJSON := `{"name":"test-theme","background":"#1a1b26","foreground":"#c0caf5","cursor":"#c0caf5","colors":["#15161e","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#a9b1d6","#414868","#f7768e","#9ece6a","#e0af68","#7aa2f7","#bb9af7","#7dcfff","#c0caf5"]}`

	// Split into chunks of ~20 chars
	chunks := []string{}
	for i := 0; i < len(themeJSON); i += 20 {
		end := i + 20
		if end > len(themeJSON) {
			end = len(themeJSON)
		}
		chunks = append(chunks, themeJSON[i:end])
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected Authorization 'Bearer test-key', got %q", auth)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)

		for _, chunk := range chunks {
			fmt.Fprint(w, sseChunk(chunk, false))
			flusher.Flush()
		}
		fmt.Fprint(w, sseChunk("", true))
		flusher.Flush()
	}))
	defer server.Close()

	var progressCalls int
	var lastTokens int
	progress := func(tokens int) {
		progressCalls++
		lastTokens = tokens
	}

	theme, err := Generate(context.Background(), server.URL, "test-key", "test-model", "a dark theme", progress)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if theme.Name != "test-theme" {
		t.Errorf("expected theme name 'test-theme', got %q", theme.Name)
	}
	if theme.Background != "#1a1b26" {
		t.Errorf("expected background '#1a1b26', got %q", theme.Background)
	}
	if theme.Foreground != "#c0caf5" {
		t.Errorf("expected foreground '#c0caf5', got %q", theme.Foreground)
	}
	if theme.Cursor != "#c0caf5" {
		t.Errorf("expected cursor '#c0caf5', got %q", theme.Cursor)
	}
	if len(theme.Colors) != 16 {
		t.Errorf("expected 16 colors, got %d", len(theme.Colors))
	}
	if theme.Colors[0] != "#15161e" {
		t.Errorf("expected colors[0] '#15161e', got %q", theme.Colors[0])
	}

	if progressCalls == 0 {
		t.Error("expected progress callback to be called at least once")
	}
	if lastTokens != progressCalls {
		t.Errorf("expected last token count %d to match progress calls %d", lastTokens, progressCalls)
	}
}

func TestGenerateConnectionRefused(t *testing.T) {
	_, err := Generate(context.Background(), "http://localhost:1", "key", "model", "test", nil)
	if err == nil {
		t.Fatal("expected error for connection refused, got nil")
	}
}

func TestGenerateTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := Generate(ctx, server.URL, "key", "model", "test", nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected error to contain 'timed out', got: %v", err)
	}
}

func TestGenerateInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)

		fmt.Fprint(w, sseChunk("this is not valid json at all", false))
		flusher.Flush()
		fmt.Fprint(w, sseChunk("", true))
		flusher.Flush()
	}))
	defer server.Close()

	_, err := Generate(context.Background(), server.URL, "key", "model", "test", nil)
	if err == nil {
		t.Fatal("expected parse error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parsing theme") {
		t.Errorf("expected error to contain 'parsing theme', got: %v", err)
	}
}

func TestGenerateAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"unauthorized"}`)
	}))
	defer server.Close()

	_, err := Generate(context.Background(), server.URL, "bad-key", "model", "test", nil)
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to contain '401', got: %v", err)
	}
}

func TestModels(t *testing.T) {
	models := Models()
	if len(models) != 3 {
		t.Fatalf("expected 3 models, got %d", len(models))
	}
	for i, m := range models {
		if m.ID == "" {
			t.Errorf("model[%d] has empty ID", i)
		}
		if m.Name == "" {
			t.Errorf("model[%d] has empty Name", i)
		}
	}
}
