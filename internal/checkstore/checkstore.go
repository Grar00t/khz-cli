package checkstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/receipt"
	"github.com/Grar00t/khz-cli/internal/term"
)

// Record is the last named check written under .khz/checks/.
type Record = receipt.Check

func pathFor(dir, name string) string {
	safe := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == 0 {
			return '_'
		}
		return r
	}, name)
	if safe == "" {
		safe = "unnamed"
	}
	return filepath.Join(dir, safe+".json")
}

// Save writes a check record atomically.
func Save(dir string, rec Record) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return atomicfile.WriteFile(pathFor(dir, rec.Name), b, 0o644)
}

// List loads all check records, sorted by name.
func List(dir string) ([]Record, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return nil, err
	}
	var out []Record
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read check %s: %w", e.Name(), err)
		}
		var rec Record
		if err := json.Unmarshal(b, &rec); err != nil {
			return nil, fmt.Errorf("parse check %s: %w", e.Name(), err)
		}
		if rec.Name == "" {
			return nil, fmt.Errorf("parse check %s: missing name", e.Name())
		}
		switch rec.Status {
		case term.OK, term.WARN, term.FAIL, term.SKIP, term.INFO:
		default:
			return nil, fmt.Errorf("parse check %s: invalid status %q", e.Name(), rec.Status)
		}
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
