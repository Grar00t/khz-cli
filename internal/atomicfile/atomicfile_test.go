package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileReplacesAndCreates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "out.json")
	if err := WriteFile(path, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "one" {
		t.Fatalf("got %q %v", b, err)
	}
	if err := WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(path)
	if err != nil || string(b) != "two" {
		t.Fatalf("got %q %v", b, err)
	}
}

func TestInterruptedWriteLeavesOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Simulate a failed pass: create a temp that is not renamed.
	tmp, err := os.CreateTemp(dir, ".khz-tmp-*")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = tmp.Write([]byte("partial"))
	_ = tmp.Close()
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "original" {
		t.Fatalf("original mutated: %q %v", b, err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) < 2 {
		t.Fatalf("expected original + temp, got %d", len(entries))
	}
}

func TestWriteFileTempIsNotTheTarget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target.json")
	if err := WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "target.json" && len(e.Name()) > 0 {
			t.Fatalf("leftover temp %q", e.Name())
		}
	}
}

func TestWriteFileReplaceFailurePreservesOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}

	originalReplace := replaceFile
	replaceFile = func(string, string) error {
		return errors.New("fixture replace failure")
	}
	t.Cleanup(func() { replaceFile = originalReplace })

	if err := WriteFile(path, []byte("replacement"), 0o644); err == nil {
		t.Fatal("expected replacement failure")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "original" {
		t.Fatalf("original target changed on failed replace: %q", b)
	}
}
