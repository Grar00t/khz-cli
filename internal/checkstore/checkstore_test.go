package checkstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListRejectsMalformedRecord(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := List(dir)
	if err == nil || !strings.Contains(err.Error(), "parse check broken.json") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestListRejectsUnknownStatus(t *testing.T) {
	dir := t.TempDir()
	raw := []byte("{\"name\":\"unit\",\"status\":\"UNKNOWN\"}\n")
	if err := os.WriteFile(filepath.Join(dir, "unit.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := List(dir)
	if err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}
