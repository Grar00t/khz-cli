package gitstate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Snapshot struct {
	InsideRepo bool     `json:"inside_repo"`
	Root       string   `json:"root,omitempty"`
	Branch     string   `json:"branch,omitempty"`
	OID        string   `json:"oid,omitempty"`
	Upstream   string   `json:"upstream,omitempty"`
	Ahead      int      `json:"ahead"`
	Behind     int      `json:"behind"`
	Detached   bool     `json:"detached"`
	Unborn     bool     `json:"unborn"`
	Clean      bool     `json:"clean"`
	Staged     int      `json:"staged"`
	Modified   int      `json:"modified"`
	Untracked  int      `json:"untracked"`
	Paths      []string `json:"paths,omitempty"`
}

var ErrGitUnavailable = errors.New("git executable not found")

func Inspect(ctx context.Context, cwd string) (Snapshot, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return Snapshot{}, ErrGitUnavailable
	}
	cmd := exec.CommandContext(ctx, git, "-C", cwd, "status", "--porcelain=v2", "-z", "--branch", "--untracked-files=all")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if strings.Contains(msg, "not a git repository") || strings.Contains(msg, "not a git repository") {
			return Snapshot{InsideRepo: false, Clean: true}, nil
		}
		return Snapshot{}, fmt.Errorf("git status: %w: %s", err, strings.TrimSpace(msg))
	}
	snap, err := ParsePorcelainV2Z(stdout.Bytes())
	if err != nil {
		return Snapshot{}, err
	}
	rootCmd := exec.CommandContext(ctx, git, "-C", cwd, "rev-parse", "--show-toplevel")
	rootOut, err := rootCmd.Output()
	if err == nil {
		snap.Root = strings.TrimSpace(string(rootOut))
	}
	return snap, nil
}

func ParsePorcelainV2Z(data []byte) (Snapshot, error) {
	s := Snapshot{InsideRepo: true, Clean: true}
	parts := bytes.Split(data, []byte{0})
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) == 0 {
			continue
		}
		rec := string(parts[i])
		switch {
		case strings.HasPrefix(rec, "# branch.oid "):
			s.OID = strings.TrimPrefix(rec, "# branch.oid ")
			if s.OID == "(initial)" {
				s.Unborn = true
				s.OID = ""
			}
		case strings.HasPrefix(rec, "# branch.head "):
			s.Branch = strings.TrimPrefix(rec, "# branch.head ")
			if s.Branch == "(detached)" {
				s.Detached = true
				s.Branch = ""
			}
		case strings.HasPrefix(rec, "# branch.upstream "):
			s.Upstream = strings.TrimPrefix(rec, "# branch.upstream ")
		case strings.HasPrefix(rec, "# branch.ab "):
			fields := strings.Fields(strings.TrimPrefix(rec, "# branch.ab "))
			if len(fields) != 2 {
				return Snapshot{}, fmt.Errorf("invalid branch.ab record %q", rec)
			}
			a, err1 := strconv.Atoi(strings.TrimPrefix(fields[0], "+"))
			b, err2 := strconv.Atoi(strings.TrimPrefix(fields[1], "-"))
			if err1 != nil || err2 != nil {
				return Snapshot{}, fmt.Errorf("invalid ahead/behind record %q", rec)
			}
			s.Ahead, s.Behind = a, b
		case strings.HasPrefix(rec, "1 "):
			fields := strings.SplitN(rec, " ", 9)
			if len(fields) != 9 {
				return Snapshot{}, fmt.Errorf("invalid ordinary record %q", rec)
			}
			applyXY(&s, fields[1])
			s.Paths = append(s.Paths, fields[8])
		case strings.HasPrefix(rec, "2 "):
			fields := strings.SplitN(rec, " ", 10)
			if len(fields) != 10 {
				return Snapshot{}, fmt.Errorf("invalid rename record %q", rec)
			}
			applyXY(&s, fields[1])
			s.Paths = append(s.Paths, fields[9])
			if i+1 >= len(parts) {
				return Snapshot{}, fmt.Errorf("rename record missing original path")
			}
			i++ // porcelain -z emits original path as the next NUL field
		case strings.HasPrefix(rec, "u "):
			fields := strings.SplitN(rec, " ", 11)
			if len(fields) != 11 {
				return Snapshot{}, fmt.Errorf("invalid unmerged record %q", rec)
			}
			applyXY(&s, fields[1])
			s.Paths = append(s.Paths, fields[10])
		case strings.HasPrefix(rec, "? "):
			s.Untracked++
			s.Clean = false
			s.Paths = append(s.Paths, strings.TrimPrefix(rec, "? "))
		case strings.HasPrefix(rec, "! "):
			// Ignored files are not requested because --ignored is not enabled.
		default:
			return Snapshot{}, fmt.Errorf("unsupported porcelain v2 record %q", rec)
		}
	}
	return s, nil
}

func applyXY(s *Snapshot, xy string) {
	if len(xy) < 2 {
		s.Clean = false
		return
	}
	if xy[0] != '.' {
		s.Staged++
		s.Clean = false
	}
	if xy[1] != '.' {
		s.Modified++
		s.Clean = false
	}
}
