package state

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Grar00t/khz-cli/internal/gitx"
)

// Root is the .khz directory for a working tree.
type Root struct {
	Project string // git root or cwd
	Dir     string // Project/.khz
}

// Locate finds the project root (git root if any, else cwd) and .khz path.
func Locate(cwd string) Root {
	project := cwd
	if r, err := gitx.FindRoot(cwd); err == nil {
		project = r
	}
	return Root{Project: project, Dir: filepath.Join(project, ".khz")}
}

func (r Root) ConfigPath() string  { return filepath.Join(r.Dir, "config.json") }
func (r Root) PolicyPath() string  { return filepath.Join(r.Dir, "policy.json") }
func (r Root) ReceiptsDir() string { return filepath.Join(r.Dir, "receipts") }
func (r Root) ChecksDir() string   { return filepath.Join(r.Dir, "checks") }

// Ensure creates .khz layout and recommends ignoring receipts in .gitignore.
func (r Root) Ensure() error {
	for _, d := range []string{r.Dir, r.ReceiptsDir(), r.ChecksDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return ensureGitignore(r.Project)
}

func ensureGitignore(project string) error {
	gi := filepath.Join(project, ".gitignore")
	const line = ".khz/receipts/"
	b, err := os.ReadFile(gi)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(gi, []byte(line+"\n"), 0o644)
		}
		return err
	}
	text := string(b)
	for _, existing := range strings.Split(text, "\n") {
		if strings.TrimSpace(existing) == line {
			return nil
		}
	}
	if !strings.HasSuffix(text, "\n") && text != "" {
		text += "\n"
	}
	text += line + "\n"
	return os.WriteFile(gi, []byte(text), 0o644)
}
