package term

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Semantic states. Color is never the only signal: these tokens are always printed.
const (
	OK   = "OK"
	WARN = "WARN"
	FAIL = "FAIL"
	SKIP = "SKIP"
	INFO = "INFO"
)

const (
	ansiReset  = "\x1b[0m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiRed    = "\x1b[31m"
	ansiGray   = "\x1b[90m"
)

// ColorConfig controls ANSI emission.
type ColorConfig struct {
	NoColor    bool
	ForceColor bool
	StdoutTTY  bool
	Getenv     func(string) string
	AllowColor bool // false when policy forbids color
}

// Enabled reports whether ANSI color should be emitted.
func (c ColorConfig) Enabled() bool {
	if c.NoColor {
		return false
	}
	if !c.AllowColor && !c.ForceColor {
		return false
	}
	getenv := c.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	if c.ForceColor {
		return true
	}
	if getenv("NO_COLOR") != "" {
		return false
	}
	if getenv("TERM") == "dumb" {
		return false
	}
	return c.StdoutTTY
}

func stateColor(state string) string {
	switch state {
	case OK:
		return ansiGreen
	case WARN:
		return ansiYellow
	case FAIL:
		return ansiRed
	case SKIP:
		return ansiGray
	default:
		return ""
	}
}

// FormatState returns STATE with optional color. The letters OK/WARN/FAIL/SKIP/INFO
// are always present.
func FormatState(state string, color bool) string {
	if !color {
		return state
	}
	c := stateColor(state)
	if c == "" {
		return state
	}
	return c + state + ansiReset
}

var (
	reOSC  = regexp.MustCompile(`\x1b\].*?(?:\x07|\x1b\\)`)
	reCSI  = regexp.MustCompile(`\x1b\[[0-9;:?]*[ -/]*[@-~]`)
	reC1   = regexp.MustCompile(`\x9b[0-9;:?]*[ -/]*[@-~]`)
	reCtrl = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
)

// Sanitize strips ANSI/control sequences from untrusted text (branch names,
// filenames, command metadata) so they cannot inject trusted UI styling.
func Sanitize(s string) string {
	s = reOSC.ReplaceAllString(s, "")
	s = reCSI.ReplaceAllString(s, "")
	s = reC1.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\x1b", "")
	s = reCtrl.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// Width returns a conservative terminal width.
func Width(getenv func(string) string, fallback int) int {
	if getenv == nil {
		getenv = os.Getenv
	}
	if c := getenv("COLUMNS"); c != "" {
		n, err := strconv.Atoi(c)
		if err == nil && n > 20 {
			return n
		}
	}
	if fallback <= 0 {
		fallback = 80
	}
	return fallback
}

// Truncate rune-safely for narrow terminals.
func Truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max-3]) + "..."
}
