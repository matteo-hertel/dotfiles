package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/chat" {
			t.Fatalf("expected /api/chat, got %s", r.URL.Path)
		}

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
