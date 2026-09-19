package app

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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

func TestStatusJSONOutsideRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	old, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	if code := Run([]string{"status", "--json"}, &out, &errBuf); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errBuf.String())
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v: %s", err, out.String())
	}
	if _, ok := got["git"]; !ok {
		t.Fatalf("missing git field: %v", got)
	}
}

func TestPolicyInitRefusesOverwrite(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	dir := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(old) })
	_ = os.Chdir(dir)
	var out, er bytes.Buffer
	if code := Run([]string{"policy", "init"}, &out, &er); code != 0 {
		t.Fatalf("first init code=%d err=%s", code, er.String())
	}
	if code := Run([]string{"policy", "init"}, &out, &er); code == 0 {
		t.Fatal("second init overwrote policy")
	}
	if _, err := os.Stat(filepath.Join(dir, ".khz", "policy.json")); err != nil {
		t.Fatal(err)
	}
}
