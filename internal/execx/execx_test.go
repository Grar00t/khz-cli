package execx

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/Grar00t/khz-cli/internal/term"
)

func TestRunZeroAndNonzero(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git missing")
	}
	var out, errb bytes.Buffer
	r := Run(context.Background(), "ver", t.TempDir(), []string{"git", "version"}, &out, &errb)
	if r.Status != term.OK || r.ExitCode != 0 {
		t.Fatalf("%+v", r)
	}
	if !strings.Contains(out.String(), "git version") {
		t.Fatalf("stdout %q", out.String())
	}

	out.Reset()
	errb.Reset()
	r = Run(context.Background(), "bad", t.TempDir(), []string{"git", "not-a-real-subcommand-khz"}, &out, &errb)
	if r.Status != term.FAIL || r.ExitCode == 0 {
		t.Fatalf("%+v stderr=%s", r, errb.String())
	}
}

func TestRunArgvWithSpaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/echo")
	}
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo missing")
	}
	var out bytes.Buffer
	r := Run(context.Background(), "echo", t.TempDir(), []string{"echo", "hello world"}, &out, ioDiscard{})
	if r.Status != term.OK {
		t.Fatalf("%+v", r)
	}
	if strings.TrimSpace(out.String()) != "hello world" {
		t.Fatalf("got %q", out.String())
	}
}

func TestRunRedactsSecrets(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git missing")
	}
	r := Run(context.Background(), "x", t.TempDir(), []string{"git", "--version", "--token", "abc"}, ioDiscard{}, ioDiscard{})
	found := false
	for i, a := range r.ArgsRedacted {
		if a == "--token" && i+1 < len(r.ArgsRedacted) && r.ArgsRedacted[i+1] == "***" {
			found = true
		}
	}
	if !found {
		t.Fatalf("redacted %#v", r.ArgsRedacted)
	}
}

func TestRunEmpty(t *testing.T) {
	r := Run(context.Background(), "x", t.TempDir(), nil, ioDiscard{}, ioDiscard{})
	if r.Status != term.FAIL || r.LookErr == "" {
		t.Fatalf("%+v", r)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
