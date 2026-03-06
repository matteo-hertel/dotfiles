package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

type modelStatus struct {
	model aigw.Model
	state string // waiting, streaming, done, failed
	tokens int
	err    error
	theme  *palette.Theme
	dur    time.Duration
}

func formatStatus(s modelStatus, spin rune) string {
	switch s.state {
	case "waiting":
		return fmt.Sprintf("  %c  %-20s waiting...", spin, s.model.Name)
	case "streaming":
		if s.tokens == 0 {
			return fmt.Sprintf("  %c  %-20s connecting...", spin, s.model.Name)
		}
		return fmt.Sprintf("  %c  %-20s streaming... %d tokens", spin, s.model.Name, s.tokens)
	case "done":
		return fmt.Sprintf("  \u2713  %-20s done (%s, %d tokens)", s.model.Name, s.dur.Truncate(time.Second), s.tokens)
	case "failed":
		errMsg := s.err.Error()
		if len(errMsg) > 60 {
			errMsg = errMsg[:60] + "..."
		}
		return fmt.Sprintf("  \u2717  %-20s failed: %s", s.model.Name, errMsg)
	default:
		return fmt.Sprintf("  %c  %-20s %s", spin, s.model.Name, s.state)
	}
}

func runAIGenerate(args []string) error {
	fs := flag.NewFlagSet("ai-generate", flag.ExitOnError)
	url := fs.String("url", aigw.DefaultBaseURL, "AI Gateway base URL")
	timeout := fs.Duration("timeout", 2*time.Minute, "Per-model timeout")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("usage: colorsync ai-generate [--url URL] [--timeout duration] <description>\n  example: colorsync ai-generate \"a warm dark theme inspired by autumn\"")
	}

	apiKey := os.Getenv("AI_GATEWAY_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("AI_GATEWAY_API_KEY environment variable is not set\n  Get your key at: https://ai-gateway.vercel.sh/setup")
	}

	description := strings.Join(remaining, " ")
	models := aigw.Models()

	fmt.Printf("Generating theme: %q\n", description)
	fmt.Printf("Racing %d models...\n", len(models))

	// Initialize status for each model
	var mu sync.Mutex
	statuses := make([]modelStatus, len(models))
	for i, m := range models {
		statuses[i] = modelStatus{model: m, state: "waiting"}
	}

	start := time.Now()
	done := make(chan struct{})

	// Launch parallel generation goroutines
	var wg sync.WaitGroup
	for i, m := range models {
		wg.Add(1)
		go func(idx int, model aigw.Model) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			defer cancel()

			mu.Lock()
			statuses[idx].state = "streaming"
			mu.Unlock()

			modelStart := time.Now()
			theme, err := aigw.Generate(ctx, *url, apiKey, model.ID, description, func(tokens int) {
				mu.Lock()
				statuses[idx].tokens = tokens
				mu.Unlock()
			})

			dur := time.Since(modelStart)

			mu.Lock()
			statuses[idx].dur = dur
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

	// Spinner goroutine for live status display
	spinChars := []rune{'\u280B', '\u2819', '\u2839', '\u2838', '\u283C', '\u2834', '\u2826', '\u2827', '\u2807', '\u280F'}
	go func() {
		iter := 0
		for {
			select {
			case <-done:
				return
			default:
				spin := spinChars[iter%len(spinChars)]

				mu.Lock()
				lines := make([]string, len(statuses))
				for i, s := range statuses {
					lines[i] = formatStatus(s, spin)
				}
				mu.Unlock()

				// Move cursor up to overwrite previous lines (skip on first iteration)
				if iter > 0 {
					fmt.Printf("\033[%dA", len(lines))
				}
				for _, line := range lines {
					fmt.Printf("\033[K%s\n", line)
				}

				iter++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Wait for all models to finish
	wg.Wait()
	close(done)

	// Small delay to let the final spinner render flush
	time.Sleep(150 * time.Millisecond)

	// Final static redraw
	mu.Lock()
	finalLines := make([]string, len(statuses))
	for i, s := range statuses {
		finalLines[i] = formatStatus(s, ' ')
	}
	mu.Unlock()

	fmt.Printf("\033[%dA", len(finalLines))
	for _, line := range finalLines {
		fmt.Printf("\033[K%s\n", line)
	}

	elapsed := time.Since(start).Truncate(time.Second)
	fmt.Printf("\nCompleted in %s\n", elapsed)

	// Collect results
	var results []int // indices of successful results
	for i, s := range statuses {
		if s.state == "done" && s.theme != nil {
			results = append(results, i)
		}
	}

	if len(results) == 0 {
		fmt.Println("\nAll models failed:")
		for _, s := range statuses {
			if s.state == "failed" {
				fmt.Printf("  %s: %v\n", s.model.Name, s.err)
			}
		}
		return fmt.Errorf("no models produced a valid theme")
	}

	// Show previews
	for num, idx := range results {
		s := statuses[idx]
		fmt.Printf("\n--- [%d] %s (%s) ---\n", num+1, s.model.Name, s.dur.Truncate(time.Second))
		preview.Render(os.Stdout, s.theme)
	}

	reader := bufio.NewReader(os.Stdin)
	var picks []int

	if len(results) == 1 {
		if confirm(reader, "Save? [y/n]: ") {
			picks = []int{0}
		}
	} else {
		answer := prompt(reader, fmt.Sprintf("Pick themes [1-%d, comma-separated, 'all', or 0 to discard]: ", len(results)))
		answer = strings.TrimSpace(answer)
		if answer == "0" {
			fmt.Println("Discarded.")
			return nil
		}
		if answer == "all" || answer == "a" {
			for i := range results {
				picks = append(picks, i)
			}
		} else {
			for _, part := range strings.Split(answer, ",") {
				n, err := strconv.Atoi(strings.TrimSpace(part))
				if err != nil || n < 1 || n > len(results) {
					fmt.Printf("Invalid selection %q, skipping.\n", strings.TrimSpace(part))
					continue
				}
				picks = append(picks, n-1)
			}
		}
	}

	if len(picks) == 0 {
		return nil
	}

	dir := palette.ThemesDir()
	if err := palette.EnsureDir(dir); err != nil {
		return err
	}
	for _, p := range picks {
		theme := statuses[results[p]].theme
		if theme == nil || theme.Name == "" {
			continue
		}
		path := filepath.Join(dir, theme.Name+".json")
		if err := theme.Save(path); err != nil {
			return err
		}
		fmt.Printf("Saved to %s\n", path)
	}

	return nil
}
