package policy

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/gitstate"
)

func TestInitDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".khz", "policy.json")
	if err := Init(path); err != nil {
		t.Fatal(err)
	}
	if err := Init(path); !errors.Is(err, atomicfile.ErrExists) {
		t.Fatalf("expected ErrExists: %v", err)
	}
}

func TestLoadRejectsUnknownAndBadVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	bad := `{"schema_version":2,"protected_branches":[],"receipt_required_for_mutations":true,"allow_color":true,"language":"en"}`
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected version failure")
	}
	unknown := `{"schema_version":1,"protected_branches":[],"receipt_required_for_mutations":true,"allow_color":true,"language":"en","extra":1}`
	if err := os.WriteFile(path, []byte(unknown), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown field failure")
	}
}

func TestCheckProtectedBranchIsInformational(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.json")
	if err := Init(path); err != nil {
		t.Fatal(err)
	}
	res := Check(path, gitstate.Snapshot{InsideRepo: true, Branch: "main"})
	if !res.Valid || len(res.Events) < 2 {
		t.Fatalf("unexpected result: %+v", res)
	}
}
