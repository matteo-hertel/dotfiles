package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhdev/dotfiles/tools/colorsync/palette"
)

const (
	p10kMarkerStart = "# --- colorsync theme start ---"
	p10kMarkerEnd   = "# --- colorsync theme end ---"
)

// GenerateP10kBlock generates a block of POWERLEVEL9K_* variable assignments
// wrapped in colorsync markers, mapping the theme palette to p10k segments.
func GenerateP10kBlock(theme *palette.Theme) string {
	bg := theme.Background
	fg := theme.Foreground
	blue := theme.Colors[4]
	red := theme.Colors[1]
	green := theme.Colors[2]
	yellow := theme.Colors[3]
	magenta := theme.Colors[5]
	brBlack := theme.Colors[8]

	var b strings.Builder
	b.WriteString(p10kMarkerStart + "\n")

	// Custom user segment: accent (blue) bg, background fg
	fmt.Fprintf(&b, "POWERLEVEL9K_CUSTOM_USER_BACKGROUND='%s'\n", blue)
	fmt.Fprintf(&b, "POWERLEVEL9K_CUSTOM_USER_FOREGROUND='%s'\n", bg)

	// Dir segment: bright black bg, foreground fg
	fmt.Fprintf(&b, "POWERLEVEL9K_DIR_BACKGROUND='%s'\n", brBlack)
	fmt.Fprintf(&b, "POWERLEVEL9K_DIR_FOREGROUND='%s'\n", fg)

	// VCS clean: green bg, background fg
	fmt.Fprintf(&b, "POWERLEVEL9K_VCS_CLEAN_BACKGROUND='%s'\n", green)
	fmt.Fprintf(&b, "POWERLEVEL9K_VCS_CLEAN_FOREGROUND='%s'\n", bg)

	// VCS modified: yellow bg, background fg
	fmt.Fprintf(&b, "POWERLEVEL9K_VCS_MODIFIED_BACKGROUND='%s'\n", yellow)
	fmt.Fprintf(&b, "POWERLEVEL9K_VCS_MODIFIED_FOREGROUND='%s'\n", bg)

	// VCS untracked: red bg, foreground fg
	fmt.Fprintf(&b, "POWERLEVEL9K_VCS_UNTRACKED_BACKGROUND='%s'\n", red)
	fmt.Fprintf(&b, "POWERLEVEL9K_VCS_UNTRACKED_FOREGROUND='%s'\n", fg)

	// Status ok: green fg (icon only, no bg)
	fmt.Fprintf(&b, "POWERLEVEL9K_STATUS_OK_FOREGROUND='%s'\n", green)

	// Date segment: bright black bg, foreground fg
	fmt.Fprintf(&b, "POWERLEVEL9K_DATE_BACKGROUND='%s'\n", brBlack)
	fmt.Fprintf(&b, "POWERLEVEL9K_DATE_FOREGROUND='%s'\n", fg)

	// Time segment: bright black bg, foreground fg
	fmt.Fprintf(&b, "POWERLEVEL9K_TIME_BACKGROUND='%s'\n", brBlack)
	fmt.Fprintf(&b, "POWERLEVEL9K_TIME_FOREGROUND='%s'\n", fg)

	// Vi mode normal: blue bg, background fg
	fmt.Fprintf(&b, "POWERLEVEL9K_VI_MODE_NORMAL_BACKGROUND='%s'\n", blue)
	fmt.Fprintf(&b, "POWERLEVEL9K_VI_MODE_NORMAL_FOREGROUND='%s'\n", bg)

	// Vi mode insert: green bg, background fg
	fmt.Fprintf(&b, "POWERLEVEL9K_VI_MODE_INSERT_BACKGROUND='%s'\n", green)
	fmt.Fprintf(&b, "POWERLEVEL9K_VI_MODE_INSERT_FOREGROUND='%s'\n", bg)

	// Vi mode visual: magenta bg, background fg
	fmt.Fprintf(&b, "POWERLEVEL9K_VI_MODE_VISUAL_BACKGROUND='%s'\n", magenta)
	fmt.Fprintf(&b, "POWERLEVEL9K_VI_MODE_VISUAL_FOREGROUND='%s'\n", bg)

	b.WriteString(p10kMarkerEnd + "\n")
	return b.String()
}

// ExportP10k injects or replaces colorsync theme variables in a Powerlevel10k
// zshtheme file. If markers already exist, the block between them is replaced.
// Otherwise the block is appended at the end of the file.
func ExportP10k(theme *palette.Theme, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading p10k config: %w", err)
	}

	block := GenerateP10kBlock(theme)
	content := string(data)

	startIdx := strings.Index(content, p10kMarkerStart)
	endIdx := strings.Index(content, p10kMarkerEnd)

	var result string
	if startIdx >= 0 && endIdx >= 0 {
		// Replace existing block (including the end marker line)
		endOfEndMarker := endIdx + len(p10kMarkerEnd)
		// Skip any trailing newline after the end marker
		if endOfEndMarker < len(content) && content[endOfEndMarker] == '\n' {
			endOfEndMarker++
		}
		result = content[:startIdx] + block + content[endOfEndMarker:]
	} else {
		// Append block at end
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			result = content + "\n" + block
		} else {
			result = content + block
		}
	}

	return os.WriteFile(path, []byte(result), 0644)
}

// P10kDefaultPath returns the default path for the Powerlevel10k zshtheme file.
func P10kDefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zshtheme")
}
