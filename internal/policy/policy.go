package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/gitx"
	"github.com/Grar00t/khz-cli/internal/term"
	"github.com/Grar00t/khz-cli/internal/version"
)

// Document is KHZ policy v1. It is local-only and is not GitHub branch protection.
type Document struct {
	SchemaVersion               int      `json:"schema_version"`
	ProtectedBranches           []string `json:"protected_branches"`
	ReceiptRequiredForMutations bool     `json:"receipt_required_for_mutations"`
	AllowColor                  bool     `json:"allow_color"`
	Language                    string   `json:"language"`
}

// Default is the document written by khz policy init.
func Default() Document {
	return Document{
		SchemaVersion:               version.SchemaVersion,
		ProtectedBranches:           []string{"main", "master"},
		ReceiptRequiredForMutations: true,
		AllowColor:                  true,
		Language:                    "en",
	}
}

// Result is the outcome of policy check.
type Result struct {
	Status                      string   `json:"status"`
	Valid                       bool     `json:"valid"`
	Path                        string   `json:"path,omitempty"`
	Branch                      string   `json:"branch,omitempty"`
	Detached                    bool     `json:"detached,omitempty"`
	ProtectedBranch             bool     `json:"protected_branch"`
	ReceiptRequiredForMutations bool     `json:"receipt_required_for_mutations"`
	AllowColor                  bool     `json:"allow_color"`
	Language                    string   `json:"language,omitempty"`
	Notes                       []string `json:"notes"`
	Errors                      []string `json:"errors,omitempty"`
}

const githubNote = "KHZ policy is local-only; it is not GitHub server-side branch protection"

// Validate returns field errors.
func Validate(d Document) []string {
	var errs []string
	if d.SchemaVersion != version.SchemaVersion {
		errs = append(errs, fmt.Sprintf("unsupported schema_version %d", d.SchemaVersion))
	}
	if d.ProtectedBranches == nil {
		errs = append(errs, "protected_branches must be an array")
	}
	lang := strings.ToLower(d.Language)
	if lang != "en" && lang != "ar" {
		errs = append(errs, "language must be en or ar")
	}
	return errs
}

// Load reads policy.json.
func Load(path string) (Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	var d Document
	if err := json.Unmarshal(b, &d); err != nil {
		return Document{}, err
	}
	return d, nil
}

// Init writes default policy if path does not exist.
func Init(path string) (Document, error) {
	if _, err := os.Stat(path); err == nil {
		return Document{}, fmt.Errorf("refusing to overwrite existing policy: %s", path)
	} else if !os.IsNotExist(err) {
		return Document{}, err
	}
	d := Default()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(d); err != nil {
		return Document{}, err
	}
	if err := atomicfile.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return Document{}, err
	}
	return d, nil
}

// Check validates the document and observes the current git branch.
func Check(path string, snap gitx.Snapshot) Result {
	res := Result{
		Notes:  []string{githubNote},
		Status: term.FAIL,
	}
	d, err := Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			res.Errors = []string{"no policy.json; run khz policy init"}
			return res
		}
		res.Errors = []string{err.Error()}
		return res
	}
	res.Path = path
	res.ReceiptRequiredForMutations = d.ReceiptRequiredForMutations
	res.AllowColor = d.AllowColor
	res.Language = d.Language
	errs := Validate(d)
	if len(errs) > 0 {
		res.Errors = errs
		return res
	}
	res.Valid = true
	res.Status = term.OK
	if snap.InRepo {
		res.Branch = snap.Head
		res.Detached = snap.Detached
		if !snap.Detached && snap.Head != "" {
			for _, b := range d.ProtectedBranches {
				if b == snap.Head {
					res.ProtectedBranch = true
					res.Status = term.WARN
					res.Notes = append(res.Notes, "current branch is listed in protected_branches (local KHZ policy)")
					break
				}
			}
		}
	}
	return res
}
