package receipt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/gitstate"
	"github.com/Grar00t/khz-cli/internal/model"
	"github.com/Grar00t/khz-cli/internal/policy"
)

const SchemaVersion = 1

type CheckEvidence struct {
	Name       string      `json:"name"`
	Status     model.State `json:"status"`
	ExitCode   int         `json:"exit_code"`
	DurationMS int64       `json:"duration_ms"`
}

type Artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

type Receipt struct {
	SchemaVersion           int                `json:"schema_version"`
	ReceiptID               string             `json:"receipt_id"`
	KHZVersion              string             `json:"khz_version"`
	StartedAt               string             `json:"started_at"`
	FinishedAt              string             `json:"finished_at"`
	DurationMS              int64              `json:"duration_ms"`
	Cwd                     string             `json:"cwd"`
	Operation               string             `json:"operation"`
	Command                 string             `json:"command"`
	ArgsRedacted            []string           `json:"args_redacted"`
	ExitCode                int                `json:"exit_code"`
	Termination             string             `json:"termination,omitempty"`
	Status                  model.State        `json:"status"`
	GitBefore               *gitstate.Snapshot `json:"git_before,omitempty"`
	GitAfter                *gitstate.Snapshot `json:"git_after,omitempty"`
	Checks                  []CheckEvidence    `json:"checks"`
	PolicyEvents            []policy.Event     `json:"policy_events"`
	Artifacts               []Artifact         `json:"artifacts"`
	SideEffectObservability string             `json:"side_effect_observability"`
	ReceiptHash             string             `json:"receipt_hash"`
}

type Verification struct {
	Valid         bool   `json:"valid"`
	ReceiptID     string `json:"receipt_id,omitempty"`
	SchemaVersion int    `json:"schema_version,omitempty"`
	StoredHash    string `json:"stored_hash,omitempty"`
	ComputedHash  string `json:"computed_hash,omitempty"`
	Error         string `json:"error,omitempty"`
}

func NewID(now time.Time) (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return now.UTC().Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(b), nil
}

func canonicalBytes(r Receipt) ([]byte, error) {
	r.ReceiptHash = ""
	return json.Marshal(r)
}

func ComputeHash(r Receipt) (string, error) {
	data, err := canonicalBytes(r)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func Finalize(r *Receipt) error {
	if r.SchemaVersion == 0 {
		r.SchemaVersion = SchemaVersion
	}
	if r.SideEffectObservability == "" {
		r.SideEffectObservability = "partial"
	}
	if r.ArgsRedacted == nil {
		r.ArgsRedacted = []string{}
	}
	if r.Checks == nil {
		r.Checks = []CheckEvidence{}
	}
	if r.PolicyEvents == nil {
		r.PolicyEvents = []policy.Event{}
	}
	if r.Artifacts == nil {
		r.Artifacts = []Artifact{}
	}
	hash, err := ComputeHash(*r)
	if err != nil {
		return err
	}
	r.ReceiptHash = hash
	return nil
}

func MarshalStored(r Receipt) ([]byte, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func Write(dir string, r Receipt) (string, error) {
	if err := Finalize(&r); err != nil {
		return "", err
	}
	if err := ValidateID(r.ReceiptID); err != nil {
		return "", err
	}
	data, err := MarshalStored(r)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, r.ReceiptID+".json")
	if err := atomicfile.WriteNew(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func ParseStrict(data []byte) (Receipt, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var r Receipt
	if err := dec.Decode(&r); err != nil {
		return Receipt{}, fmt.Errorf("decode receipt: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Receipt{}, errors.New("multiple JSON values")
		}
		return Receipt{}, fmt.Errorf("trailing receipt data: %w", err)
	}
	if r.SchemaVersion != SchemaVersion {
		return Receipt{}, fmt.Errorf("unsupported receipt schema_version %d", r.SchemaVersion)
	}
	if err := ValidateID(r.ReceiptID); err != nil {
		return Receipt{}, err
	}
	if r.SideEffectObservability != "partial" && r.SideEffectObservability != "full" {
		return Receipt{}, fmt.Errorf("invalid side_effect_observability %q", r.SideEffectObservability)
	}
	if err := validateReceipt(r); err != nil {
		return Receipt{}, err
	}
	return r, nil
}

func validateReceipt(r Receipt) error {
	if r.KHZVersion == "" || r.StartedAt == "" || r.FinishedAt == "" || r.Cwd == "" || r.Operation == "" || r.Command == "" {
		return errors.New("receipt missing required field")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.StartedAt); err != nil {
		return fmt.Errorf("invalid started_at: %w", err)
	}
	if _, err := time.Parse(time.RFC3339Nano, r.FinishedAt); err != nil {
		return fmt.Errorf("invalid finished_at: %w", err)
	}
	if r.DurationMS < 0 {
		return errors.New("duration_ms cannot be negative")
	}
	if r.ArgsRedacted == nil || r.Checks == nil || r.PolicyEvents == nil || r.Artifacts == nil {
		return errors.New("receipt arrays must be present")
	}
	switch r.Status {
	case model.OK, model.Warn, model.Fail, model.Skip, model.Info:
	default:
		return fmt.Errorf("invalid receipt status %q", r.Status)
	}
	if len(r.ReceiptHash) != 64 {
		return errors.New("invalid receipt_hash length")
	}
	if _, err := hex.DecodeString(r.ReceiptHash); err != nil {
		return errors.New("receipt_hash must be hexadecimal")
	}
	return nil
}

func VerifyBytes(data []byte) Verification {
	r, err := ParseStrict(data)
	if err != nil {
		return Verification{Valid: false, Error: err.Error()}
	}
	computed, err := ComputeHash(r)
	if err != nil {
		return Verification{Valid: false, ReceiptID: r.ReceiptID, Error: err.Error()}
	}
	v := Verification{ReceiptID: r.ReceiptID, SchemaVersion: r.SchemaVersion, StoredHash: r.ReceiptHash, ComputedHash: computed}
	if !strings.EqualFold(r.ReceiptHash, computed) {
		v.Error = "receipt hash mismatch"
		return v
	}
	v.Valid = true
	return v
}

func VerifyFile(path string) Verification {
	data, err := os.ReadFile(path)
	if err != nil {
		return Verification{Valid: false, Error: err.Error()}
	}
	return VerifyBytes(data)
}

func ValidateID(id string) error {
	if id == "" || id == "." || id == ".." || filepath.Base(id) != id ||
		strings.ContainsRune(id, '/') || strings.ContainsRune(id, '\\') ||
		strings.ContainsRune(id, '\x00') || strings.ContainsRune(id, '\r') || strings.ContainsRune(id, '\n') {
		return errors.New("invalid receipt id")
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune("._-", r)) {
			return errors.New("invalid receipt id")
		}
	}
	return nil
}

func PathForID(dir, id string) (string, error) {
	id = strings.TrimSuffix(id, ".json")
	if err := ValidateID(id); err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".json"), nil
}

func List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		if ValidateID(id) == nil {
			ids = append(ids, id)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	return ids, nil
}
