package term

import (
	"strings"
	"testing"
)

func TestSanitizeStripsCSIFromUntrustedText(t *testing.T) {
	in := "main\x1b[31mSECRET\x1b[0m"
	got := Sanitize(in)
	if strings.Contains(got, "\x1b") {
		t.Fatalf("escape survived: %q", got)
	}
	if strings.Contains(got, "[31m") {
		t.Fatalf("CSI payload leaked: %q", got)
	}
	if !strings.Contains(got, "SECRET") || !strings.Contains(got, "main") {
		t.Fatalf("payload lost: %q", got)
	}
}

func TestSanitizeStripsOSCAndC0(t *testing.T) {
	in := "file\x1b]0;title\x07name\x00.bin"
	got := Sanitize(in)
	if strings.Contains(got, "\x1b") || strings.Contains(got, "\x00") || strings.Contains(got, "\x07") {
		t.Fatalf("control survived: %q", got)
	}
}

func TestSanitizeFlattensNewlines(t *testing.T) {
	got := Sanitize("a\nb\rc")
	if strings.ContainsAny(got, "\n\r") {
		t.Fatalf("newline survived: %q", got)
	}
}

func TestColorDisabledByNO_COLOR(t *testing.T) {
	c := ColorConfig{
		StdoutTTY:  true,
		AllowColor: true,
		Getenv: func(k string) string {
			if k == "NO_COLOR" {
				return "1"
			}
			return ""
		},
	}
	if c.Enabled() {
		t.Fatal("NO_COLOR must disable color")
	}
}

func TestColorDisabledByDumbTERM(t *testing.T) {
	c := ColorConfig{
		StdoutTTY:  true,
		AllowColor: true,
		Getenv: func(k string) string {
			if k == "TERM" {
				return "dumb"
			}
			return ""
		},
	}
	if c.Enabled() {
		t.Fatal("TERM=dumb must disable color")
	}
}

func TestColorDisabledWhenNotTTY(t *testing.T) {
	c := ColorConfig{StdoutTTY: false, AllowColor: true, Getenv: func(string) string { return "" }}
	if c.Enabled() {
		t.Fatal("non-TTY must disable color")
	}
}

func TestForceColorOverridesNonTTY(t *testing.T) {
	c := ColorConfig{ForceColor: true, StdoutTTY: false, AllowColor: true, Getenv: func(string) string { return "" }}
	if !c.Enabled() {
		t.Fatal("--color should force ANSI")
	}
}

func TestNoColorWinsOverForce(t *testing.T) {
	c := ColorConfig{NoColor: true, ForceColor: true, StdoutTTY: true, AllowColor: true, Getenv: func(string) string { return "" }}
	if c.Enabled() {
		t.Fatal("--no-color must win")
	}
}

func TestFormatStateAlwaysContainsToken(t *testing.T) {
	for _, st := range []string{OK, WARN, FAIL, SKIP, INFO} {
		plain := FormatState(st, false)
		colored := FormatState(st, true)
		if plain != st {
			t.Fatalf("plain %q", plain)
		}
		if !strings.Contains(colored, st) {
			t.Fatalf("colored missing token: %q", colored)
		}
	}
}

func TestTruncate(t *testing.T) {
	if Truncate("abc", 10) != "abc" {
		t.Fatal("short")
	}
	got := Truncate("abcdefghij", 6)
	if got != "abc..." {
		t.Fatalf("got %q", got)
	}
}
