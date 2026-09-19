package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/gitstate"
	"github.com/Grar00t/khz-cli/internal/model"
)

const SchemaVersion = 1

type Policy struct {
	SchemaVersion               int      `json:"schema_version"`
	ProtectedBranches           []string `json:"protected_branches"`
	ReceiptRequiredForMutations bool     `json:"receipt_required_for_mutations"`
	AllowColor                  bool     `json:"allow_color"`
	Language                    string   `json:"language"`
}

type Event struct {
	State  model.State `json:"state"`
	Kind   string      `json:"kind"`
	Detail string      `json:"detail"`
}

type CheckResult struct {
	Valid  bool        `json:"valid"`
	State  model.State `json:"state"`
	Events []Event     `json:"events"`
	Policy *Policy     `json:"policy,omitempty"`
}

func Default() Policy {
	return Policy{SchemaVersion: SchemaVersion, ProtectedBranches: []string{"main", "master"}, ReceiptRequiredForMutations: true, AllowColor: true, Language: "en"}
}

func Init(path string) error {
	p := Default()
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return atomicfile.WriteNew(path, data, 0o600)
}

func Load(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p Policy
	if err := dec.Decode(&p); err != nil {
		return Policy{}, fmt.Errorf("decode policy: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Policy{}, errors.New("multiple JSON values in policy")
		}
		return Policy{}, fmt.Errorf("trailing policy data: %w", err)
	}
	if err := Validate(p); err != nil {
		return Policy{}, err
	}
	return p, nil
}

func Validate(p Policy) error {
	if p.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported policy schema_version %d", p.SchemaVersion)
	}
	if p.Language != "en" && p.Language != "ar" {
		return fmt.Errorf("language must be en or ar")
	}
	seen := map[string]bool{}
	for _, b := range p.ProtectedBranches {
		if strings.TrimSpace(b) == "" {
			return errors.New("protected branch cannot be empty")
		}
		if strings.ContainsAny(b, "\r\n\x00\x1b") {
			return errors.New("protected branch contains control characters")
		}
		if seen[b] {
			return fmt.Errorf("duplicate protected branch %q", b)
		}
		seen[b] = true
	}
	return nil
}

func Check(path string, git gitstate.Snapshot) CheckResult {
	p, err := Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CheckResult{Valid: false, State: model.Warn, Events: []Event{{State: model.Warn, Kind: "policy_missing", Detail: "policy not initialized"}}}
		}
		return CheckResult{Valid: false, State: model.Fail, Events: []Event{{State: model.Fail, Kind: "policy_invalid", Detail: err.Error()}}}
	}
	events := []Event{{State: model.OK, Kind: "schema", Detail: "policy schema valid"}}
	if git.InsideRepo && !git.Detached {
		for _, b := range p.ProtectedBranches {
			if git.Branch == b {
				events = append(events, Event{State: model.Info, Kind: "protected_branch", Detail: "current branch is locally marked protected; this is not GitHub server protection"})
				break
			}
		}
	}
	return CheckResult{Valid: true, State: model.OK, Events: events, Policy: &p}
}
