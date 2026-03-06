# AI Generate Command — Design

## Goal

Add an `ai-generate` command to colorsync that uses Apple Foundation Models (on-device) to generate a complete terminal color theme from a natural language description.

## Architecture

A Swift helper binary (`colorsync-ai`) uses Foundation Models' guided generation to produce structured theme JSON from a text prompt. The Go CLI shells out to this binary, parses the JSON output, previews the result, and saves it.

```
./colorsync ai-generate "warm dark theme inspired by autumn"
        │
        ▼
  cmd/ai_generate.go  (Go)
    │ exec.Command("colorsync-ai", prompt)
    ▼
  tools/colorsync-ai/  (Swift CLI)
    │ FoundationModels @Generable + LanguageModelSession
    ▼
  JSON to stdout → parsed as palette.Theme
    │
    ▼
  preview → confirm → save
```

## Key Decisions

- **Swift helper binary** over Python SDK or AppleScript — full API access, compiled, no runtime deps
- **Guided generation** via `@Generable` — guarantees valid structured output matching Theme schema
- **Separate Swift package** — keeps Go tool stdlib-only, Swift binary built independently
- **Binary discovery** — looks adjacent to Go binary first, then in PATH
