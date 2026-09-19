package receipt

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Grar00t/khz-cli/internal/gitx"
	"github.com/Grar00t/khz-cli/internal/term"
	"github.com/Grar00t/khz-cli/internal/version"
)

func sample(now time.Time) Receipt {
	gb := gitx.ParsePorcelainV2Z([]byte("# branch.oid abcdef0123456789\x00# branch.head main\x00"))
	return Receipt{
		SchemaVersion:           version.SchemaVersion,
		ReceiptID:               "khz_test_01",
		KHZVersion:              version.Version,
		StartedAt:               now.UTC().Format(time.RFC3339Nano),
		FinishedAt:              now.UTC().Add(time.Second).Format(time.RFC3339Nano),
		DurationMS:              1000,
		CWD:                     "/tmp/proj",
		Operation:               "check",
		Command:                 "git",
		ArgsRedacted:            []string{"--password", "***"},
		ExitCode:                0,
		Status:                  term.OK,
		GitBefore:               &gb,
		GitAfter:                &gb,
		Checks:                  []Check{{Name: "unit", Status: term.OK, ExitCode: 0, DurationMS: 1000, FinishedAt: now.UTC().Format(time.RFC3339Nano), Command: "git"}},
		PolicyEvents:            []PolicyEvent{},
		Artifacts:               []string{},
		SideEffectObservability: SideEffectPartial,
	}
}

func TestHashStableAndSealRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 19, 1, 0, 0, 0, time.UTC)
	r := sample(now)
	h1, err := Hash(r)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := Hash(r)
	if err != nil || h1 != h2 {
		t.Fatalf("unstable hash %s %s %v", h1, h2, err)
	}
	if err := Seal(&r); err != nil {
		t.Fatal(err)
	}
	if r.ReceiptHash != h1 {
		t.Fatal("seal mismatch")
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	vr := VerifyBytes(raw)
	if !vr.OK {
		t.Fatalf("%+v", vr)
	}
}

func TestVerifyRejectsTamper(t *testing.T) {
	r := sample(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if err := Seal(&r); err != nil {
		t.Fatal(err)
	}
	r.ExitCode = 99
	raw, _ := json.Marshal(r)
	vr := VerifyBytes(raw)
	if vr.OK || vr.Status != term.FAIL || !strings.Contains(vr.Reason, "mismatch") {
		t.Fatalf("%+v", vr)
	}
}

func TestVerifyRejectsMalformed(t *testing.T) {
	vr := VerifyBytes([]byte("{not json"))
	if vr.OK || !strings.Contains(vr.Reason, "malformed") {
		t.Fatalf("%+v", vr)
	}
}

func TestVerifyRejectsUnsupportedSchema(t *testing.T) {
	r := sample(time.Now())
	r.SchemaVersion = 99
	if err := Seal(&r); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(r)
	vr := VerifyBytes(raw)
	if vr.OK || !strings.Contains(vr.Reason, "unsupported") {
		t.Fatalf("%+v", vr)
	}
}

func TestWriteAtomicAndListShow(t *testing.T) {
	dir := t.TempDir()
	r := sample(time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC))
	path, err := Write(dir, &r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	ids, err := List(dir)
	if err != nil || len(ids) != 1 || ids[0] != r.ReceiptID {
		t.Fatalf("%v %v", ids, err)
	}
	loaded, raw, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ReceiptID != r.ReceiptID {
		t.Fatal(loaded.ReceiptID)
	}
	vr := VerifyBytes(raw)
	if !vr.OK {
		t.Fatalf("%+v", vr)
	}
	got := Resolve(dir, r.ReceiptID)
	if got != path {
		t.Fatalf("resolve %q %q", got, path)
	}
}

func TestHashIgnoresStoredHashField(t *testing.T) {
	r := sample(time.Unix(0, 0).UTC())
	h, _ := Hash(r)
	r.ReceiptHash = "deadbeef"
	h2, _ := Hash(r)
	if h != h2 {
		t.Fatal("hash must ignore receipt_hash contents")
	}
}

func TestCanonicalDoesNotHTMLEscape(t *testing.T) {
	r := sample(time.Unix(1, 0).UTC())
	r.Command = "a<b>"
	b, err := marshalCanonical(r)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("\\u003c")) {
		t.Fatalf("html escaped: %s", b)
	}
}

func TestNewID(t *testing.T) {
	now := time.Date(2026, 9, 19, 4, 5, 6, 0, time.UTC)
	id, err := NewID(now, func(b []byte) (int, error) {
		for i := range b {
			b[i] = byte(i + 1)
		}
		return len(b), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id, "khz_20260919T040506Z_") {
		t.Fatalf("%s", id)
	}
}
