package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Grar00t/khz-cli/internal/exitcode"
	"github.com/Grar00t/khz-cli/internal/term"
	"github.com/Grar00t/khz-cli/internal/version"
)

func runCLI(t *testing.T, cwd string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errb bytes.Buffer
	env := Env{
		Args:      append([]string{"khz"}, args...),
		Stdout:    &out,
		Stderr:    &errb,
		CWD:       cwd,
		Getenv:    func(string) string { return "" },
		StdoutTTY: false,
	}
	code = Main(env)
	return out.String(), errb.String(), code
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func repo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git missing")
	}
	dir := t.TempDir()
	git(t, dir, "init")
	git(t, dir, "config", "user.email", "t@example.com")
	git(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "README")
	git(t, dir, "commit", "-m", "init")
	return dir
}

func TestVersion(t *testing.T) {
	out, _, code := runCLI(t, t.TempDir(), "version")
	if code != 0 || !strings.Contains(out, version.Version) {
		t.Fatalf("code=%d out=%q", code, out)
	}
	out, _, code = runCLI(t, t.TempDir(), "version", "--json")
	if code != 0 {
		t.Fatal(code)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(out), &m); err != nil || m["version"] != version.Version {
		t.Fatalf("%v %s", err, out)
	}
}

func TestUsageExit(t *testing.T) {
	_, errb, code := runCLI(t, t.TempDir())
	if code != exitcode.Usage || !strings.Contains(errb, "not a terminal emulator") {
		t.Fatalf("code=%d err=%q", code, errb)
	}
	_, _, code = runCLI(t, t.TempDir(), "nope")
	if code != exitcode.Usage {
		t.Fatal(code)
	}
}

func TestHelp(t *testing.T) {
	out, _, code := runCLI(t, t.TempDir(), "--help")
	if code != 0 || !strings.Contains(out, "not a shell replacement") {
		t.Fatalf("%d %s", code, out)
	}
}

func TestGitStatusJSONAndOutside(t *testing.T) {
	_, _, code := runCLI(t, t.TempDir(), "git", "status", "--json")
	if code != exitcode.Fail {
		t.Fatalf("outside repo code %d", code)
	}
	dir := repo(t)
	out, _, code := runCLI(t, dir, "git", "status", "--json")
	if code != 0 {
		t.Fatalf("code %d %s", code, out)
	}
	var snap map[string]any
	if err := json.Unmarshal([]byte(out), &snap); err != nil {
		t.Fatal(err)
	}
	if snap["in_repo"] != true || snap["clean"] != true {
		t.Fatalf("%v", snap)
	}
}

func TestStatusJSON(t *testing.T) {
	dir := repo(t)
	out, _, code := runCLI(t, dir, "status", "--json")
	if code != 0 {
		t.Fatalf("%d %s", code, out)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["khz_version"] != version.Version {
		t.Fatalf("%v", payload)
	}
}

func TestRunAndCheckAndBoard(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git missing")
	}
	dir := repo(t)
	out, errb, code := runCLI(t, dir, "run", "--", "git", "version")
	if code != 0 {
		t.Fatalf("run %d out=%s err=%s", code, out, errb)
	}
	out, errb, code = runCLI(t, dir, "check", "--name", "unit", "--", "git", "version")
	if code != 0 {
		t.Fatalf("check %d %s %s", code, out, errb)
	}
	out, _, code = runCLI(t, dir, "board", "--json")
	if code != 0 {
		t.Fatalf("board %d %s", code, out)
	}
	var b map[string]any
	if err := json.Unmarshal([]byte(out), &b); err != nil {
		t.Fatal(err)
	}
	// Ensure() writes .gitignore / .khz which dirties git → WARN, so REVIEW is the
	// honest decision. The named check must still appear as OK.
	rows, _ := b["rows"].([]any)
	found := false
	for _, r := range rows {
		m := r.(map[string]any)
		if m["name"] == "unit" && m["status"] == term.OK {
			found = true
		}
	}
	if !found {
		t.Fatalf("board missing real check: %s", out)
	}
	if b["decision"] == "FAIL" {
		t.Fatalf("unit-only should not FAIL: %s", out)
	}
	human, _, _ := runCLI(t, dir, "--no-color", "board")
	if !strings.Contains(human, "OK") || !strings.Contains(human, "unit") {
		t.Fatalf("human board: %s", human)
	}
	if strings.Contains(human, "\x1b") {
		t.Fatal("color leaked with --no-color")
	}

	_, errb, code = runCLI(t, dir, "check", "--name", "bad", "--", "git", "definitely-not-a-command")
	if code != exitcode.Fail {
		t.Fatalf("expected fail %d %s", code, errb)
	}
	out, _, code = runCLI(t, dir, "board", "--json")
	var b2 map[string]any
	_ = json.Unmarshal([]byte(out), &b2)
	if b2["decision"] != "FAIL" {
		t.Fatalf("failing check must not yield PASS/PROCEED: %v", b2["decision"])
	}
}

func TestRunRequiresDashDash(t *testing.T) {
	_, errb, code := runCLI(t, t.TempDir(), "run", "git", "version")
	if code != exitcode.Usage || !strings.Contains(errb, "--") {
		t.Fatalf("%d %s", code, errb)
	}
}

func TestArgvSpaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo")
	}
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo")
	}
	dir := repo(t)
	out, errb, code := runCLI(t, dir, "run", "--json", "--", "echo", "hello world")
	if code != 0 {
		t.Fatalf("%d %s %s", code, out, errb)
	}
	if !strings.Contains(errb, "hello world") {
		t.Fatalf("child stdout should go to stderr when --json: %q", errb)
	}
}

func TestReceiptVerifyAcceptAndReject(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git")
	}
	dir := repo(t)
	_, _, code := runCLI(t, dir, "run", "--", "git", "version")
	if code != 0 {
		t.Fatal(code)
	}
	out, _, code := runCLI(t, dir, "receipt", "list", "--json")
	if code != 0 {
		t.Fatal(out)
	}
	var list struct {
		Receipts []string `json:"receipts"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil || len(list.Receipts) == 0 {
		t.Fatalf("%v %s", err, out)
	}
	id := list.Receipts[0]
	show, _, code := runCLI(t, dir, "receipt", "show", id)
	if code != 0 || !strings.Contains(show, `"receipt_id"`) {
		t.Fatalf("show %d %s", code, show)
	}
	_, _, code = runCLI(t, dir, "receipt", "verify", id)
	if code != 0 {
		t.Fatal("verify clean")
	}
	path := filepath.Join(dir, ".khz", "receipts", id+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(b), `"exit_code": 0`, `"exit_code": 7`, 1)
	if tampered == string(b) {
		t.Fatal("could not tamper")
	}
	if err := os.WriteFile(path, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errb, code := runCLI(t, dir, "receipt", "verify", id)
	if code == 0 {
		t.Fatalf("tamper accepted: %s", errb)
	}
	bad := filepath.Join(dir, ".khz", "receipts", "malformed.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, code = runCLI(t, dir, "receipt", "verify", "malformed")
	if code == 0 {
		t.Fatal("malformed accepted")
	}
}

func TestPolicyInitAndCheck(t *testing.T) {
	dir := repo(t)
	out, errb, code := runCLI(t, dir, "policy", "init")
	if code != 0 {
		t.Fatalf("%d %s %s", code, out, errb)
	}
	_, errb, code = runCLI(t, dir, "policy", "init")
	if code == 0 || !strings.Contains(errb, "overwrite") {
		t.Fatalf("overwrite %d %s", code, errb)
	}
	out, _, code = runCLI(t, dir, "policy", "check", "--json")
	if code != 0 {
		t.Fatal(out)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatal(err)
	}
	if res["valid"] != true {
		t.Fatalf("%v", res)
	}
	notes, _ := res["notes"].([]any)
	joined := ""
	for _, n := range notes {
		joined += n.(string)
	}
	if !strings.Contains(joined, "not GitHub") {
		t.Fatalf("disclaimer missing: %v", res)
	}
}

func TestNoColorAndNonTTY(t *testing.T) {
	dir := repo(t)
	out, _, _ := runCLI(t, dir, "--no-color", "status")
	if strings.Contains(out, "\x1b") {
		t.Fatalf("ansi: %q", out)
	}
	var buf bytes.Buffer
	env := Env{
		Args:      []string{"khz", "status"},
		Stdout:    &buf,
		Stderr:    ioDiscard{},
		CWD:       dir,
		StdoutTTY: false,
		Getenv:    func(string) string { return "" },
	}
	Main(env)
	if strings.Contains(buf.String(), "\x1b") {
		t.Fatal("non-tty colored")
	}
	buf.Reset()
	env.Getenv = func(k string) string {
		if k == "NO_COLOR" {
			return "1"
		}
		return ""
	}
	env.StdoutTTY = true
	env.Args = []string{"khz", "status"}
	Main(env)
	if strings.Contains(buf.String(), "\x1b") {
		t.Fatal("NO_COLOR ignored")
	}
}

func TestSecretRedactionInReceipt(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo")
	}
	dir := repo(t)
	out, errb, code := runCLI(t, dir, "run", "--json", "--", "echo", "--token", "super-secret-value")
	if code != 0 {
		t.Fatalf("run %d out=%s err=%s", code, out, errb)
	}
	listOut, _, _ := runCLI(t, dir, "receipt", "list", "--json")
	var list struct {
		Receipts []string `json:"receipts"`
	}
	_ = json.Unmarshal([]byte(listOut), &list)
	if len(list.Receipts) == 0 {
		t.Fatal("no receipts")
	}
	show, _, _ := runCLI(t, dir, "receipt", "show", list.Receipts[0])
	if strings.Contains(show, "super-secret-value") {
		t.Fatal("secret leaked into receipt")
	}
	if !strings.Contains(show, "***") {
		t.Fatalf("expected redaction: %s", show)
	}
}

func TestANSIInjectionInBoard(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git")
	}
	dir := repo(t)
	// branch names cannot contain many control chars, so inject via a check name
	_, _, code := runCLI(t, dir, "check", "--name", "x\x1b[31mHACK", "--", "git", "version")
	if code != 0 {
		t.Fatal(code)
	}
	out, _, _ := runCLI(t, dir, "--color", "board")
	if strings.Contains(out, "[31m") {
		t.Fatalf("injected CSI: %q", out)
	}
}

func TestDoctor(t *testing.T) {
	dir := repo(t)
	out, _, code := runCLI(t, dir, "doctor", "--json")
	if code != 0 {
		t.Fatalf("%d %s", code, out)
	}
	if !strings.Contains(out, "partial") {
		t.Fatalf("observability note missing: %s", out)
	}
}

func TestLangARJSONKeysEnglish(t *testing.T) {
	dir := repo(t)
	out, _, code := runCLI(t, dir, "--lang", "ar", "status", "--json")
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out, `"git"`) {
		t.Fatal("json keys must stay English")
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
