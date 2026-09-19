package gitstate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParsePorcelainSpacesUnicodeAndNUL(t *testing.T) {
	data := []byte("# branch.oid abc\x00# branch.head main\x001 .M N... 100644 100644 100644 a b file with spaces.txt\x00? مرحبا.txt\x00")
	got, err := ParsePorcelainV2Z(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Clean || got.Modified != 1 || got.Untracked != 1 {
		t.Fatalf("unexpected snapshot: %+v", got)
	}
	if len(got.Paths) != 2 || got.Paths[0] != "file with spaces.txt" || got.Paths[1] != "مرحبا.txt" {
		t.Fatalf("paths not preserved: %#v", got.Paths)
	}
}

func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func TestInspectTemporaryRepositories(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Git behavior is still exercised in CI, but Windows temp-path behavior can differ.
	}
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-b", "main")

	s, err := Inspect(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if !s.InsideRepo || !s.Unborn || s.Branch != "main" {
		t.Fatalf("unborn snapshot: %+v", s)
	}

	gitCmd(t, dir, "config", "user.name", "KHZ Test")
	gitCmd(t, dir, "config", "user.email", "khz@example.invalid")
	if err := os.WriteFile(filepath.Join(dir, "space name.txt"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "مرحبا.txt"), []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}

	s, err = Inspect(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Untracked != 2 || s.Clean {
		t.Fatalf("untracked snapshot: %+v", s)
	}

	gitCmd(t, dir, "add", ".")
	s, err = Inspect(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Staged != 2 || s.Clean {
		t.Fatalf("staged snapshot: %+v", s)
	}
	gitCmd(t, dir, "commit", "-m", "initial")
	if err := os.WriteFile(filepath.Join(dir, "space name.txt"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err = Inspect(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Modified != 1 {
		t.Fatalf("modified snapshot: %+v", s)
	}

	gitCmd(t, dir, "checkout", "--detach", "HEAD")
	s, err = Inspect(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Detached {
		t.Fatalf("expected detached: %+v", s)
	}
}

func TestInspectOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	s, err := Inspect(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.InsideRepo {
		t.Fatalf("expected outside repo: %+v", s)
	}
}

func TestParseAheadBehind(t *testing.T) {
	data := []byte("# branch.oid abc\x00# branch.head main\x00# branch.upstream origin/main\x00# branch.ab +3 -2\x00")
	s, err := ParsePorcelainV2Z(data)
	if err != nil {
		t.Fatal(err)
	}
	if s.Upstream != "origin/main" || s.Ahead != 3 || s.Behind != 2 {
		t.Fatalf("unexpected: %+v", s)
	}
}
