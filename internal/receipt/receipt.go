package receipt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/gitx"
	"github.com/Grar00t/khz-cli/internal/version"
)

const currentSchema = version.SchemaVersion

// SideEffectPartial is the only observability KHZ claims for arbitrary commands.
const SideEffectPartial = "partial"

// Receipt is the versioned evidence object written under .khz/receipts/.
type Receipt struct {
	SchemaVersion           int            `json:"schema_version"`
	ReceiptID               string         `json:"receipt_id"`
	KHZVersion              string         `json:"khz_version"`
	StartedAt               string         `json:"started_at"`
	FinishedAt              string         `json:"finished_at"`
	DurationMS              int64          `json:"duration_ms"`
	CWD                     string         `json:"cwd"`
	Operation               string         `json:"operation"`
	Command                 string         `json:"command"`
	ArgsRedacted            []string       `json:"args_redacted"`
	ExitCode                int            `json:"exit_code"`
	Status                  string         `json:"status"`
	GitBefore               *gitx.Snapshot `json:"git_before"`
	GitAfter                *gitx.Snapshot `json:"git_after"`
	Checks                  []Check        `json:"checks"`
	PolicyEvents            []PolicyEvent  `json:"policy_events"`
	Artifacts               []string       `json:"artifacts"`
	SideEffectObservability string         `json:"side_effect_observability"`
	ReceiptHash             string         `json:"receipt_hash"`
}

// Check is a named check recorded on a receipt.
type Check struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	ExitCode   int    `json:"exit_code"`
	DurationMS int64  `json:"duration_ms"`
	FinishedAt string `json:"finished_at"`
	Command    string `json:"command"`
}

// PolicyEvent is a local policy observation (not GitHub branch protection).
type PolicyEvent struct {
	Code    string `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewID returns a time-prefixed unique id. now and readRand are injectable.
func NewID(now time.Time, readRand func([]byte) (int, error)) (string, error) {
	if readRand == nil {
		readRand = rand.Read
	}
	var b [8]byte
	if _, err := readRand(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("khz_%s_%s", now.UTC().Format("20060102T150405Z"), hex.EncodeToString(b[:])), nil
}

func marshalCanonical(r Receipt) ([]byte, error) {
	r.ReceiptHash = ""
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte{'\n'}), nil
}

// Hash returns SHA-256 hex of the canonical JSON with receipt_hash empty.
func Hash(r Receipt) (string, error) {
	b, err := marshalCanonical(r)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// Seal fills ReceiptHash.
func Seal(r *Receipt) error {
	h, err := Hash(*r)
	if err != nil {
		return err
	}
	r.ReceiptHash = h
	return nil
}

// VerifyResult is the outcome of khz receipt verify.
type VerifyResult struct {
	OK       bool   `json:"ok"`
	Status   string `json:"status"`
	Reason   string `json:"reason,omitempty"`
	ID       string `json:"receipt_id,omitempty"`
	Path     string `json:"path,omitempty"`
	Hash     string `json:"receipt_hash,omitempty"`
	Computed string `json:"computed_hash,omitempty"`
}

// VerifyBytes validates a receipt document.
func VerifyBytes(raw []byte) VerifyResult {
	var r Receipt
	if err := json.Unmarshal(raw, &r); err != nil {
		return VerifyResult{OK: false, Status: "FAIL", Reason: "malformed: " + err.Error()}
	}
	if r.SchemaVersion == 0 && r.ReceiptID == "" {
		return VerifyResult{OK: false, Status: "FAIL", Reason: "malformed: missing schema_version and receipt_id"}
	}
	if r.SchemaVersion != currentSchema {
		return VerifyResult{OK: false, Status: "FAIL", Reason: fmt.Sprintf("unsupported schema_version %d", r.SchemaVersion)}
	}
	if r.ReceiptID == "" || r.ReceiptHash == "" {
		return VerifyResult{OK: false, Status: "FAIL", Reason: "malformed: missing receipt_id or receipt_hash"}
	}
	want, err := Hash(r)
	if err != nil {
		return VerifyResult{OK: false, Status: "FAIL", Reason: err.Error()}
	}
	if !strings.EqualFold(want, r.ReceiptHash) {
		return VerifyResult{OK: false, Status: "FAIL", Reason: "receipt_hash mismatch (tampered or non-canonical)", ID: r.ReceiptID, Hash: r.ReceiptHash, Computed: want}
	}
	return VerifyResult{OK: true, Status: "OK", ID: r.ReceiptID, Hash: r.ReceiptHash, Computed: want}
}

// Write atomically writes a sealed receipt to dir/<id>.json.
func Write(dir string, r *Receipt) (string, error) {
	if r.ReceiptHash == "" {
		if err := Seal(r); err != nil {
			return "", err
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, r.ReceiptID+".json")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return "", err
	}
	if err := atomicfile.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// ReadFile loads a receipt from path.
func ReadFile(path string) (Receipt, []byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Receipt{}, nil, err
	}
	var r Receipt
	if err := json.Unmarshal(b, &r); err != nil {
		return Receipt{}, b, err
	}
	return r, b, nil
}

// List returns receipt ids (filenames without .json) sorted.
func List(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			ids = append(ids, strings.TrimSuffix(name, ".json"))
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// Resolve maps an id or path to a file under dir.
func Resolve(dir, idOrPath string) string {
	if strings.ContainsRune(idOrPath, os.PathSeparator) || strings.HasSuffix(idOrPath, ".json") {
		if filepath.IsAbs(idOrPath) {
			return idOrPath
		}
		if _, err := os.Stat(idOrPath); err == nil {
			return idOrPath
		}
	}
	base := idOrPath
	if !strings.HasSuffix(base, ".json") {
		base += ".json"
	}
	return filepath.Join(dir, base)
}
