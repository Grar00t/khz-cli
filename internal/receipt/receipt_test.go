package receipt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Grar00t/khz-cli/internal/model"
)

func vector() Receipt {
	return Receipt{
		SchemaVersion:           SchemaVersion,
		ReceiptID:               "20260101T000000.000000000Z-abcdef123456",
		KHZVersion:              "0.1.0",
		StartedAt:               "2026-01-01T00:00:00Z",
		FinishedAt:              "2026-01-01T00:00:01Z",
		DurationMS:              1000,
		Cwd:                     "/tmp/example",
		Operation:               "check",
		Command:                 "example",
		ArgsRedacted:            []string{"--token=<redacted>"},
		ExitCode:                0,
		Status:                  model.OK,
		SideEffectObservability: "partial",
	}
}

func TestHashVectorDeterministic(t *testing.T) {
	r := vector()
	if err := Finalize(&r); err != nil {
		t.Fatal(err)
	}
	const expected = "422d2dc1d1b1516cd486f93af8281866f894f8f8acfe9efa83f5709d987a32d7"
	if r.ReceiptHash != expected {
		t.Fatalf("hash=%s", r.ReceiptHash)
	}
	first := r.ReceiptHash
	if err := Finalize(&r); err != nil {
		t.Fatal(err)
	}
	if r.ReceiptHash != first {
		t.Fatalf("hash changed: %s != %s", r.ReceiptHash, first)
	}
}

func TestVerifyUntouchedAndTampered(t *testing.T) {
	r := vector()
	if err := Finalize(&r); err != nil {
		t.Fatal(err)
	}
	data, err := MarshalStored(r)
	if err != nil {
		t.Fatal(err)
	}
	if v := VerifyBytes(data); !v.Valid {
		t.Fatalf("valid receipt rejected: %+v", v)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}
	obj["cwd"] = "/tampered"
	tampered, _ := json.Marshal(obj)
	if v := VerifyBytes(tampered); v.Valid {
		t.Fatal("tampered receipt accepted")
	}
}

func TestMalformedAndUnsupportedVersion(t *testing.T) {
	if v := VerifyBytes([]byte("{")); v.Valid {
		t.Fatal("malformed receipt accepted")
	}
	missing := []byte(`{"schema_version":1,"receipt_id":"abc","receipt_hash":"0000000000000000000000000000000000000000000000000000000000000000","side_effect_observability":"partial"}`)
	if v := VerifyBytes(missing); v.Valid || !strings.Contains(v.Error, "missing required") {
		t.Fatalf("missing-field receipt accepted: %+v", v)
	}
	r := vector()
	r.SchemaVersion = 99
	r.ReceiptHash = strings.Repeat("0", 64)
	data, _ := json.Marshal(r)
	if v := VerifyBytes(data); v.Valid || !strings.Contains(v.Error, "unsupported") {
		t.Fatalf("unexpected: %+v", v)
	}
}

func TestWriteRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	r := vector()
	if _, err := Write(dir, r); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(dir, r); err == nil {
		t.Fatal("overwrite accepted")
	}
}

func TestListAndTraversal(t *testing.T) {
	dir := t.TempDir()
	r := vector()
	if _, err := Write(dir, r); err != nil {
		t.Fatal(err)
	}
	ids, err := List(dir)
	if err != nil || len(ids) != 1 || ids[0] != r.ReceiptID {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	if _, err := PathForID(dir, "../secret"); err == nil {
		t.Fatal("path traversal accepted")
	}
}

func TestNewIDValid(t *testing.T) {
	id, err := NewID(time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateID(id); err != nil {
		t.Fatal(err)
	}
}

func TestStoredReceiptPermissionsAndAtomicPresence(t *testing.T) {
	dir := t.TempDir()
	r := vector()
	path, err := Write(dir, r)
	if err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o077 != 0 {
		t.Fatalf("permissions too broad: %v", st.Mode().Perm())
	}
	if filepath.Ext(path) != ".json" {
		t.Fatalf("unexpected path %q", path)
	}
}
