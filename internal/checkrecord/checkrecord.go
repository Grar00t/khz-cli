package checkrecord

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/model"
)

const SchemaVersion = 1

type Record struct {
	SchemaVersion int         `json:"schema_version"`
	Name          string      `json:"name"`
	Status        model.State `json:"status"`
	Command       string      `json:"command"`
	ArgsRedacted  []string    `json:"args_redacted"`
	Cwd           string      `json:"cwd"`
	StartedAt     string      `json:"started_at"`
	FinishedAt    string      `json:"finished_at"`
	DurationMS    int64       `json:"duration_ms"`
	ExitCode      int         `json:"exit_code"`
	ReceiptID     string      `json:"receipt_id"`
}

func fileName(name string) (string, error) {
	if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\x00\r\n\x1b") {
		return "", errors.New("invalid check name")
	}
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	sum := sha256.Sum256([]byte(name))
	return b.String() + "-" + hex.EncodeToString(sum[:4]) + ".json", nil
}

func Write(dir string, r Record) (string, error) {
	if r.SchemaVersion == 0 {
		r.SchemaVersion = SchemaVersion
	}
	name, err := fileName(r.Name)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	path := filepath.Join(dir, name)
	if err := atomicfile.WriteReplace(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func LoadAll(dir string) ([]Record, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	records := []Record{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		var r Record
		if err := dec.Decode(&r); err != nil {
			return nil, fmt.Errorf("decode check %s: %w", e.Name(), err)
		}
		if r.SchemaVersion != SchemaVersion {
			return nil, fmt.Errorf("unsupported check schema %d", r.SchemaVersion)
		}
		records = append(records, r)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	return records, nil
}
