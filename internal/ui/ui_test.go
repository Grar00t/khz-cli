package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Grar00t/khz-cli/internal/model"
)

func TestSanitizeControlSequences(t *testing.T) {
	got := Sanitize("branch\x1b[31mRED\r\n")
	if strings.ContainsRune(got, '\x1b') || strings.ContainsRune(got, '\r') || strings.ContainsRune(got, '\n') {
		t.Fatalf("control sequence survived: %q", got)
	}
	if !strings.Contains(got, "\\x1B") {
		t.Fatalf("escape not rendered safely: %q", got)
	}
}

func TestNonTTYNoColor(t *testing.T) {
	var b bytes.Buffer
	if ColorEnabled(&b, false, true) {
		t.Fatal("buffer must not be treated as TTY")
	}
	Line(&b, model.OK, "source", "clean", false)
	if strings.Contains(b.String(), "\x1b[") {
		t.Fatal("ANSI emitted to redirected output")
	}
}

func TestNoColorEnvironment(t *testing.T) {
	old := os.Getenv("NO_COLOR")
	t.Cleanup(func() { _ = os.Setenv("NO_COLOR", old) })
	_ = os.Setenv("NO_COLOR", "1")
	if ColorEnabled(os.Stdout, false, true) {
		t.Fatal("NO_COLOR ignored")
	}
}

func TestLineSanitizesUntrustedState(t *testing.T) {
	var b bytes.Buffer
	Line(&b, model.State("OK\x1b[31m"), "name", "detail", false)
	if strings.ContainsRune(b.String(), '\x1b') {
		t.Fatalf("ANSI injection survived: %q", b.String())
	}
}

func TestSanitizeBidiControls(t *testing.T) {
	got := Sanitize("safe\u202Etxt")
	if strings.ContainsRune(got, '\u202E') {
		t.Fatalf("bidi override survived: %q", got)
	}
	if !strings.Contains(got, "\\u202E") {
		t.Fatalf("bidi override not made visible: %q", got)
	}
}
