package gitx

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseUnborn(t *testing.T) {
	s := ParsePorcelainV2Z([]byte("# branch.oid (initial)\x00# branch.head master\x00"))
	if !s.InRepo || !s.Initial || s.Head != "master" || s.OID != "" {
		t.Fatalf("%+v", s)
	}
	if s.Clean {
		t.Fatal("unborn is not clean")
	}
}

func TestParseClean(t *testing.T) {
	s := ParsePorcelainV2Z([]byte("# branch.oid 53390e20638c18f2845a63f85ffa26b033b8c89b\x00# branch.head master\x00"))
	if !s.Clean || s.ShortOID != "53390e2" || s.Head != "master" || s.Detached {
		t.Fatalf("%+v", s)
	}
}

func TestParseMixedSpaces(t *testing.T) {
	raw := []byte("# branch.oid 53390e20638c18f2845a63f85ffa26b033b8c89b\x00# branch.head master\x001 .M N... 100644 100644 100644 78981922613b2afb6025042ff6bd878ac1994e85 78981922613b2afb6025042ff6bd878ac1994e85 file with spaces.txt\x001 AM N... 000000 100644 100644 0000000000000000000000000000000000000000 d905d9da82c97264ab6f4920e20242e088850ce9 tracked.txt\x00? untracked.dat\x00")
	s := ParsePorcelainV2Z(raw)
	if !reflect.DeepEqual(s.Modified, []string{"file with spaces.txt", "tracked.txt"}) {
		t.Fatalf("modified %#v", s.Modified)
	}
	if !reflect.DeepEqual(s.Staged, []string{"tracked.txt"}) {
		t.Fatalf("staged %#v", s.Staged)
	}
	if !reflect.DeepEqual(s.Untracked, []string{"untracked.dat"}) {
		t.Fatalf("untracked %#v", s.Untracked)
	}
	if s.Clean {
		t.Fatal("dirty")
	}
}

func TestParseDetached(t *testing.T) {
	s := ParsePorcelainV2Z([]byte("# branch.oid 53390e20638c18f2845a63f85ffa26b033b8c89b\x00# branch.head (detached)\x00"))
	if !s.Detached || s.Head != "" || !s.Clean {
		t.Fatalf("%+v", s)
	}
}

func TestParseRenameNUL(t *testing.T) {
	raw := []byte("# branch.oid ca58fbafb7c44c6ff45a065914084cd4c362a419\x00# branch.head master\x002 R. N... 100644 100644 100644 34eaebd7e7b7fed59f301ffab126e68f469ea1ab 34eaebd7e7b7fed59f301ffab126e68f469ea1ab R100 renamed.txt\x00tracked.txt\x00")
	s := ParsePorcelainV2Z(raw)
	if len(s.Renames) != 1 || s.Renames[0].From != "tracked.txt" || s.Renames[0].To != "renamed.txt" {
		t.Fatalf("renames %#v", s.Renames)
	}
	if !reflect.DeepEqual(s.Staged, []string{"renamed.txt"}) {
		t.Fatalf("staged %#v", s.Staged)
	}
}

func TestParseAheadBehind(t *testing.T) {
	s := ParsePorcelainV2Z([]byte("# branch.oid abcdef0deadbeef\x00# branch.head main\x00# branch.upstream origin/main\x00# branch.ab +2 -3\x00"))
	if s.Upstream != "origin/main" || s.Ahead == nil || *s.Ahead != 2 || s.Behind == nil || *s.Behind != 3 {
		t.Fatalf("%+v", s)
	}
}

func gitOk(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, errb.String())
	}
}

func TestProbeTempRepos(t *testing.T) {
	gitOk(t)
	ctx := context.Background()

	t.Run("outside", func(t *testing.T) {
		dir := t.TempDir()
		s := Probe(ctx, dir)
		if s.InRepo {
			t.Fatalf("expected not in repo: %+v", s)
		}
		if s.Error != ErrNotRepository.Error() {
			t.Fatalf("error %q", s.Error)
		}
	})

	t.Run("unborn", func(t *testing.T) {
		dir := t.TempDir()
		git(t, dir, "init")
		s := Probe(ctx, dir)
		if !s.InRepo || !s.Initial {
			t.Fatalf("%+v", s)
		}
	})

	t.Run("clean modified staged untracked spaces unicode", func(t *testing.T) {
		dir := t.TempDir()
		git(t, dir, "init")
		git(t, dir, "config", "user.email", "t@example.com")
		git(t, dir, "config", "user.name", "t")
		if err := os.WriteFile(filepath.Join(dir, "file with spaces.txt"), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "عربي.txt"), []byte("b\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "add", "-A")
		git(t, dir, "commit", "-m", "init")
		s := Probe(ctx, dir)
		if !s.Clean {
			t.Fatalf("want clean %+v", s)
		}

		if err := os.WriteFile(filepath.Join(dir, "file with spaces.txt"), []byte("a\nc\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "untracked.dat"), []byte("d\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("e\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "add", "staged.txt")
		s = Probe(ctx, dir)
		if s.Clean {
			t.Fatal("want dirty")
		}
		foundSpace, foundUnicode, foundUntracked, foundStaged := false, false, false, false
		for _, p := range s.Modified {
			if p == "file with spaces.txt" {
				foundSpace = true
			}
		}
		for _, p := range s.Untracked {
			if p == "untracked.dat" {
				foundUntracked = true
			}
		}
		for _, p := range s.Staged {
			if p == "staged.txt" {
				foundStaged = true
			}
		}
		// unicode file is unmodified tracked; confirm it is not listed
		for _, p := range append(append(s.Modified, s.Untracked...), s.Staged...) {
			if p == "عربي.txt" {
				foundUnicode = true
			}
		}
		if !foundSpace || !foundUntracked || !foundStaged {
			t.Fatalf("space=%v untracked=%v staged=%v snap=%+v", foundSpace, foundUntracked, foundStaged, s)
		}
		if foundUnicode {
			t.Fatal("unmodified unicode file should not appear")
		}

		if err := os.WriteFile(filepath.Join(dir, "عربي.txt"), []byte("b\nchanged\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		s = Probe(ctx, dir)
		foundUnicode = false
		for _, p := range s.Modified {
			if p == "عربي.txt" {
				foundUnicode = true
			}
		}
		if !foundUnicode {
			t.Fatalf("unicode modified missing: %+v", s)
		}
	})

	t.Run("detached", func(t *testing.T) {
		dir := t.TempDir()
		git(t, dir, "init")
		git(t, dir, "config", "user.email", "t@example.com")
		git(t, dir, "config", "user.name", "t")
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "add", "a.txt")
		git(t, dir, "commit", "-m", "c")
		git(t, dir, "checkout", "--detach", "HEAD")
		s := Probe(ctx, dir)
		if !s.Detached {
			t.Fatalf("%+v", s)
		}
	})
}

func TestFindRoot(t *testing.T) {
	gitOk(t)
	dir := t.TempDir()
	git(t, dir, "init")
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := FindRoot(sub)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := filepath.EvalSymlinks(root)
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if _, err := FindRoot(t.TempDir()); err == nil {
		t.Fatal("expected error outside repo")
	}
}
