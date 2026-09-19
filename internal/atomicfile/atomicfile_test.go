package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteNewAndNoOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x", "receipt.json")
	if err := WriteNew(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteNew(path, []byte("second"), 0o600); !errors.Is(err, ErrExists) {
		t.Fatalf("expected ErrExists, got %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first" {
		t.Fatalf("file overwritten: %q", got)
	}
}

func TestWriteReplace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.json")
	if err := WriteReplace(path, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteReplace(path, []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "b" {
		t.Fatalf("got %q", got)
	}
}
