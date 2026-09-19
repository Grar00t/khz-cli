package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Grar00t/khz-cli/internal/gitx"
	"github.com/Grar00t/khz-cli/internal/term"
)

func TestInitRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "policy.json")
	if _, err := Init(p); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(p); err == nil || !strings.Contains(err.Error(), "overwrite") {
		t.Fatalf("err %v", err)
	}
}

func TestCheckMissing(t *testing.T) {
	res := Check(filepath.Join(t.TempDir(), "policy.json"), gitx.Snapshot{})
	if res.Status != term.FAIL || res.Valid {
		t.Fatalf("%+v", res)
	}
}

func TestCheckValidAndProtected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "policy.json")
	if _, err := Init(p); err != nil {
		t.Fatal(err)
	}
	snap := gitx.Snapshot{InRepo: true, Head: "main"}
	res := Check(p, snap)
	if !res.Valid || res.Status != term.WARN || !res.ProtectedBranch {
		t.Fatalf("%+v", res)
	}
	if !strings.Contains(strings.Join(res.Notes, " "), "not GitHub") {
		t.Fatalf("missing github disclaimer: %+v", res)
	}
	snap.Head = "topic"
	res = Check(p, snap)
	if res.Status != term.OK || res.ProtectedBranch {
		t.Fatalf("%+v", res)
	}
}

func TestValidateLanguage(t *testing.T) {
	d := Default()
	d.Language = "xx"
	if errs := Validate(d); len(errs) == 0 {
		t.Fatal("expected language error")
	}
}

func TestCheckInvalidJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "policy.json")
	if err := os.WriteFile(p, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := Check(p, gitx.Snapshot{})
	if res.Status != term.FAIL {
		t.Fatalf("%+v", res)
	}
}
