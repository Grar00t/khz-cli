package gitx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Snapshot is the machine-readable Git state KHZ records. It is parsed from
// `git status --porcelain=v2 -z --branch`, not from human git status.
type Snapshot struct {
	InRepo    bool     `json:"in_repo"`
	Error     string   `json:"error,omitempty"`
	OID       string   `json:"oid,omitempty"`
	ShortOID  string   `json:"short_oid,omitempty"`
	Initial   bool     `json:"initial"`
	Head      string   `json:"head,omitempty"`
	Detached  bool     `json:"detached"`
	Upstream  string   `json:"upstream,omitempty"`
	Ahead     *int     `json:"ahead,omitempty"`
	Behind    *int     `json:"behind,omitempty"`
	Staged    []string `json:"staged"`
	Modified  []string `json:"modified"`
	Untracked []string `json:"untracked"`
	Unmerged  []string `json:"unmerged"`
	Renames   []Rename `json:"renames"`
	Clean     bool     `json:"clean"`
}

// Rename is a porcelain v2 rename/copy record.
type Rename struct {
	From string `json:"from"`
	To   string `json:"to"`
	XY   string `json:"xy"`
}

var (
	ErrNotRepository = errors.New("not a git repository")
	ErrGitMissing    = errors.New("git executable not found")
)

func emptySnap() Snapshot {
	return Snapshot{
		Staged:    []string{},
		Modified:  []string{},
		Untracked: []string{},
		Unmerged:  []string{},
		Renames:   []Rename{},
	}
}

// Probe returns a snapshot for cwd using argv-based git (no shell).
func Probe(ctx context.Context, cwd string) Snapshot {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}
	git, err := exec.LookPath("git")
	if err != nil {
		s := emptySnap()
		s.Error = ErrGitMissing.Error()
		return s
	}
	cmd := exec.CommandContext(ctx, git, "status", "--porcelain=v2", "-z", "--branch")
	cmd.Dir = cwd
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = runErr.Error()
		}
		s := emptySnap()
		if strings.Contains(strings.ToLower(msg), "not a git repository") {
			s.Error = ErrNotRepository.Error()
		} else {
			s.Error = msg
		}
		return s
	}
	s := ParsePorcelainV2Z(stdout.Bytes())
	s.InRepo = true
	s.Clean = isClean(s)
	return s
}

func isClean(s Snapshot) bool {
	return s.InRepo && !s.Initial &&
		len(s.Staged) == 0 && len(s.Modified) == 0 &&
		len(s.Untracked) == 0 && len(s.Unmerged) == 0 &&
		len(s.Renames) == 0
}

// ParsePorcelainV2Z parses NUL-terminated porcelain v2 --branch output.
func ParsePorcelainV2Z(data []byte) Snapshot {
	s := emptySnap()
	parts := bytes.Split(data, []byte{0})
	for i := 0; i < len(parts); i++ {
		rec := string(parts[i])
		if rec == "" {
			continue
		}
		switch {
		case strings.HasPrefix(rec, "# branch.oid "):
			oid := strings.TrimPrefix(rec, "# branch.oid ")
			if oid == "(initial)" {
				s.Initial = true
			} else {
				s.OID = oid
				if len(oid) >= 7 {
					s.ShortOID = oid[:7]
				} else {
					s.ShortOID = oid
				}
			}
		case strings.HasPrefix(rec, "# branch.head "):
			head := strings.TrimPrefix(rec, "# branch.head ")
			if head == "(detached)" {
				s.Detached = true
			} else {
				s.Head = head
			}
		case strings.HasPrefix(rec, "# branch.upstream "):
			s.Upstream = strings.TrimPrefix(rec, "# branch.upstream ")
		case strings.HasPrefix(rec, "# branch.ab "):
			parseAB(strings.TrimPrefix(rec, "# branch.ab "), &s)
		case strings.HasPrefix(rec, "1 "):
			parseOrdinary(rec, &s)
		case strings.HasPrefix(rec, "2 "):
			orig := ""
			if i+1 < len(parts) {
				orig = string(parts[i+1])
				i++
			}
			parseRename(rec, orig, &s)
		case strings.HasPrefix(rec, "u "):
			path := restPath(rec, 11)
			if path != "" {
				s.Unmerged = append(s.Unmerged, path)
			}
		case strings.HasPrefix(rec, "? "):
			s.Untracked = append(s.Untracked, rec[2:])
		}
	}
	if s.OID != "" || s.Initial || s.Head != "" || s.Detached {
		s.InRepo = true
	}
	s.Clean = isClean(s)
	return s
}

func parseAB(ab string, s *Snapshot) {
	fields := strings.Fields(ab)
	if len(fields) < 2 {
		return
	}
	if strings.HasPrefix(fields[0], "+") {
		if n, err := strconv.Atoi(fields[0][1:]); err == nil {
			s.Ahead = &n
		}
	}
	if strings.HasPrefix(fields[1], "-") {
		if n, err := strconv.Atoi(fields[1][1:]); err == nil {
			s.Behind = &n
		}
	}
}

func parseOrdinary(rec string, s *Snapshot) {
	fields := strings.SplitN(rec, " ", 9)
	if len(fields) < 9 {
		return
	}
	xy := fields[1]
	path := fields[8]
	if len(xy) < 2 {
		return
	}
	if xy[0] != '.' {
		s.Staged = appendUnique(s.Staged, path)
	}
	if xy[1] != '.' {
		s.Modified = appendUnique(s.Modified, path)
	}
}

func parseRename(rec, orig string, s *Snapshot) {
	fields := strings.SplitN(rec, " ", 10)
	if len(fields) < 10 {
		return
	}
	xy := fields[1]
	dest := fields[9]
	s.Renames = append(s.Renames, Rename{From: orig, To: dest, XY: xy})
	if len(xy) >= 1 && xy[0] != '.' {
		s.Staged = appendUnique(s.Staged, dest)
	}
	if len(xy) >= 2 && xy[1] != '.' {
		s.Modified = appendUnique(s.Modified, dest)
	}
}

func restPath(rec string, n int) string {
	fields := strings.SplitN(rec, " ", n)
	if len(fields) < n {
		return ""
	}
	return fields[n-1]
}

func appendUnique(ss []string, v string) []string {
	for _, x := range ss {
		if x == v {
			return ss
		}
	}
	return append(ss, v)
}

// FindRoot walks parents looking for a .git file or directory.
func FindRoot(cwd string) (string, error) {
	dir, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}
	for {
		p := filepath.Join(dir, ".git")
		if st, err := os.Stat(p); err == nil && (st.IsDir() || st.Mode().IsRegular()) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%w", ErrNotRepository)
		}
		dir = parent
	}
}
