package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Grar00t/khz-cli/internal/model"
)

var colors = map[model.State]string{
	model.OK:   "\x1b[32m",
	model.Warn: "\x1b[33m",
	model.Fail: "\x1b[31m",
	model.Skip: "\x1b[90m",
}

const reset = "\x1b[0m"

func IsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && (info.Mode()&os.ModeCharDevice) != 0
}

func ColorEnabled(w io.Writer, noColor bool, allowColor bool) bool {
	if noColor || !allowColor || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return IsTTY(w)
}

func Sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) || isBidiControl(r) {
			if r <= 0xff {
				fmt.Fprintf(&b, "\\x%02X", r)
			} else {
				fmt.Fprintf(&b, "\\u%04X", r)
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isBidiControl(r rune) bool {
	switch r {
	case 0x200e, 0x200f, 0x202a, 0x202b, 0x202c, 0x202d, 0x202e, 0x2066, 0x2067, 0x2068, 0x2069:
		return true
	default:
		return false
	}
}

func terminalWidth() int {
	if n, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && n >= 24 {
		return n
	}
	return 80
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	if max <= 3 {
		return strings.Repeat(".", max)
	}
	r := []rune(s)
	return string(r[:max-3]) + "..."
}

func Line(w io.Writer, state model.State, name, detail string, color bool) {
	name = Sanitize(name)
	detail = Sanitize(detail)
	width := terminalWidth()
	const fixed = 5 + 2 + 12 + 2
	if width > fixed {
		detail = truncate(detail, width-fixed)
	}
	marker := fmt.Sprintf("%-5s", Sanitize(string(state)))
	if color {
		if c := colors[state]; c != "" {
			marker = c + marker + reset
		}
	}
	fmt.Fprintf(w, "%s  %-12s  %s\n", marker, truncate(name, 12), detail)
}

func Text(lang, key string) string {
	ar := map[string]string{
		"decision":        "القرار",
		"receipt":         "الإيصال",
		"not_repo":        "ليس داخل مستودع Git",
		"clean":           "المصدر نظيف",
		"dirty":           "توجد تغييرات محلية",
		"policy_missing":  "السياسة غير موجودة",
		"policy_valid":    "السياسة صالحة",
		"checks_missing":  "لا توجد فحوصات مسجلة",
		"doctor_title":    "تشخيص KHZ",
		"status_title":    "حالة KHZ",
		"board_title":     "لوحة قرار KHZ",
		"verified":        "الإيصال صحيح",
		"invalid_receipt": "الإيصال غير صالح",
	}
	if lang == "ar" {
		if v, ok := ar[key]; ok {
			return v
		}
	}
	en := map[string]string{
		"decision":        "DECISION",
		"receipt":         "RECEIPT",
		"not_repo":        "not inside a Git repository",
		"clean":           "source clean",
		"dirty":           "local changes present",
		"policy_missing":  "policy not initialized",
		"policy_valid":    "policy valid",
		"checks_missing":  "no recorded checks",
		"doctor_title":    "KHZ doctor",
		"status_title":    "KHZ status",
		"board_title":     "KHZ decision board",
		"verified":        "receipt verified",
		"invalid_receipt": "receipt invalid",
	}
	return en[key]
}
