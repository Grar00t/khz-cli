package board

import (
	"fmt"
	"strings"

	"github.com/Grar00t/khz-cli/internal/i18n"
	"github.com/Grar00t/khz-cli/internal/term"
)

// Row is one board line. State is always a textual token.
type Row struct {
	State  string `json:"status"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

// Board is the human decision surface.
type Board struct {
	Headline string `json:"headline"`
	Rows     []Row  `json:"rows"`
	Decision string `json:"decision"`
	Receipt  string `json:"receipt,omitempty"`
}

// Decide computes the overall decision. FAIL is never hidden under PROCEED.
func Decide(rows []Row) string {
	hasFail, hasWarn, hasOK := false, false, false
	measured := 0
	for _, r := range rows {
		switch r.State {
		case term.FAIL:
			hasFail = true
			measured++
		case term.WARN:
			hasWarn = true
			measured++
		case term.OK:
			hasOK = true
			measured++
		case term.SKIP, term.INFO:
			// not a pass
		}
	}
	if hasFail {
		return "FAIL"
	}
	if hasWarn {
		return "REVIEW"
	}
	if hasOK && measured > 0 {
		return "PROCEED"
	}
	return "INCOMPLETE"
}

// Render writes a plain, no-blink board. Color is optional; tokens always appear.
func Render(b Board, color bool, lang string, width int) string {
	if width < 40 {
		width = 40
	}
	var lines []string
	lines = append(lines, "KHZ  "+term.Sanitize(b.Headline))
	lines = append(lines, "")
	for _, r := range linesRows(b.Rows, color, lang, width) {
		lines = append(lines, r)
	}
	lines = append(lines, "")
	decLabel := i18n.T(lang, "decision")
	recLabel := i18n.T(lang, "receipt")
	lines = append(lines, fmt.Sprintf("%-9s %s", decLabel, b.Decision))
	rec := b.Receipt
	if rec == "" {
		rec = "-"
	}
	lines = append(lines, fmt.Sprintf("%-9s %s", recLabel, term.Sanitize(rec)))
	return strings.Join(lines, "\n") + "\n"
}

func linesRows(rows []Row, color bool, lang string, width int) []string {
	var out []string
	for _, row := range rows {
		st := term.FormatState(row.State, color)
		name := term.Sanitize(i18n.T(lang, row.Name))
		if name == row.Name {
			name = term.Sanitize(row.Name)
		}
		detail := term.Sanitize(row.Detail)
		// pad name to 10 runes using spaces (ASCII names)
		padded := pad(name, 10)
		line := fmt.Sprintf("%s    %s  %s", padVisible(st, 4), padded, detail)
		if width > 0 {
			line = term.Truncate(line, width)
		}
		out = append(out, line)
	}
	return out
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padVisible(s string, n int) string {
	plain := term.Sanitize(s)
	// if s contains ANSI, pad based on token width 4
	if len(plain) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(plain))
}
