package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var out, err bytes.Buffer
	code := Run([]string{"version"}, &out, &err)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if got := out.String(); got != "khz "+Version+"\n" {
		t.Fatalf("unexpected output %q", got)
	}
	if err.Len() != 0 {
		t.Fatalf("unexpected stderr %q", err.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var out, err bytes.Buffer
	code := Run([]string{"nope"}, &out, &err)
	if code != 2 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(err.String(), "unknown command") {
		t.Fatalf("unexpected stderr %q", err.String())
	}
}
