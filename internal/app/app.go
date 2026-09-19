package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Grar00t/khz-cli/internal/atomicfile"
	"github.com/Grar00t/khz-cli/internal/board"
	"github.com/Grar00t/khz-cli/internal/checkrecord"
	"github.com/Grar00t/khz-cli/internal/doctor"
	"github.com/Grar00t/khz-cli/internal/gitstate"
	"github.com/Grar00t/khz-cli/internal/model"
	"github.com/Grar00t/khz-cli/internal/policy"
	"github.com/Grar00t/khz-cli/internal/receipt"
	"github.com/Grar00t/khz-cli/internal/runner"
	"github.com/Grar00t/khz-cli/internal/ui"
	"github.com/Grar00t/khz-cli/internal/workspace"
)

const Version = "0.1.0-dev"

type globalOptions struct {
	Lang    string
	NoColor bool
}

type statusOutput struct {
	KHZVersion   string             `json:"khz_version"`
	Root         string             `json:"root"`
	State        model.State        `json:"state"`
	Git          gitstate.Snapshot  `json:"git"`
	Policy       policy.CheckResult `json:"policy"`
	ReceiptCount int                `json:"receipt_count"`
	CheckCount   int                `json:"check_count"`
}

func Run(args []string, stdout, stderr io.Writer) int {
	return RunIO(args, nil, stdout, stderr)
}

func RunIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, rest, err := parseGlobal(args)
	if err != nil {
		fmt.Fprintf(stderr, "khz: %v\n", err)
		return 2
	}
	if len(rest) == 0 {
		printHelp(stdout)
		return 0
	}
	ctx := context.Background()
	cmd := rest[0]
	argv := rest[1:]
	switch cmd {
	case "version":
		if len(argv) != 0 {
			return usageError(stderr, "version takes no arguments")
		}
		fmt.Fprintf(stdout, "khz %s\n", Version)
		return 0
	case "help", "--help", "-h":
		printHelp(stdout)
		return 0
	case "doctor":
		return commandDoctor(argv, opts, stdout, stderr)
	case "status":
		return commandStatus(ctx, argv, opts, stdout, stderr)
	case "run":
		return commandRun(ctx, argv, opts, stdin, stdout, stderr, "run", "")
	case "check":
		return commandCheck(ctx, argv, opts, stdin, stdout, stderr)
	case "board":
		return commandBoard(ctx, argv, opts, stdout, stderr)
	case "receipt":
		return commandReceipt(ctx, argv, opts, stdout, stderr)
	case "policy":
		return commandPolicy(ctx, argv, opts, stdout, stderr)
	case "git":
		return commandGit(ctx, argv, opts, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "khz: unknown command %q\n", ui.Sanitize(cmd))
		fmt.Fprintln(stderr, "Run 'khz help' for usage.")
		return 2
	}
}

func parseGlobal(args []string) (globalOptions, []string, error) {
	opts := globalOptions{Lang: "en"}
	for len(args) > 0 {
		switch {
		case args[0] == "--no-color":
			opts.NoColor = true
			args = args[1:]
		case args[0] == "--lang":
			if len(args) < 2 {
				return opts, nil, errors.New("--lang requires en or ar")
			}
			opts.Lang = args[1]
			args = args[2:]
		case strings.HasPrefix(args[0], "--lang="):
			opts.Lang = strings.TrimPrefix(args[0], "--lang=")
			args = args[1:]
		default:
			if opts.Lang != "en" && opts.Lang != "ar" {
				return opts, nil, fmt.Errorf("unsupported language %q", opts.Lang)
			}
			return opts, args, nil
		}
	}
	if opts.Lang != "en" && opts.Lang != "ar" {
		return opts, nil, fmt.Errorf("unsupported language %q", opts.Lang)
	}
	return opts, args, nil
}

func usageError(stderr io.Writer, msg string) int {
	fmt.Fprintf(stderr, "khz: %s\n", msg)
	return 2
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "KHZ - Human Decision CLI")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "KHZ is not a terminal emulator or shell replacement.")
	fmt.Fprintln(w, "It is a structured decision and receipt layer for existing command-line workflows.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  khz [--lang en|ar] [--no-color] <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  version")
	fmt.Fprintln(w, "  doctor [--json]")
	fmt.Fprintln(w, "  status [--json]")
	fmt.Fprintln(w, "  run -- <command> [args...]")
	fmt.Fprintln(w, "  check --name <name> -- <command> [args...]")
	fmt.Fprintln(w, "  board [--json]")
	fmt.Fprintln(w, "  receipt list [--json]")
	fmt.Fprintln(w, "  receipt show <id>")
	fmt.Fprintln(w, "  receipt verify <id> [--json]")
	fmt.Fprintln(w, "  policy init")
	fmt.Fprintln(w, "  policy check [--json]")
	fmt.Fprintln(w, "  git status [--json]")
}

func jsonFlag(args []string) (bool, []string) {
	out := make([]string, 0, len(args))
	jsonMode := false
	for _, a := range args {
		if a == "--json" {
			jsonMode = true
		} else {
			out = append(out, a)
		}
	}
	return jsonMode, out
}

func writeJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(value)
}

func commandDoctor(args []string, opts globalOptions, stdout, stderr io.Writer) int {
	jsonMode, rest := jsonFlag(args)
	if len(rest) != 0 {
		return usageError(stderr, "doctor accepts only --json")
	}
	r := doctor.Build(Version)
	if jsonMode {
		if err := writeJSON(stdout, r); err != nil {
			fmt.Fprintf(stderr, "khz: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, ui.Text(opts.Lang, "doctor_title"))
	color := ui.ColorEnabled(stdout, opts.NoColor, true)
	ui.Line(stdout, model.OK, "khz", r.KHZVersion, color)
	ui.Line(stdout, model.Info, "runtime", r.GoVersion+" "+r.OS+"/"+r.Arch, color)
	for _, tool := range r.Tools {
		state, detail := model.Skip, "not available"
		if tool.Available {
			state, detail = model.OK, tool.Path
		}
		ui.Line(stdout, state, tool.Name, detail, color)
	}
	return 0
}

func inspectGit(ctx context.Context, cwd string) (gitstate.Snapshot, error) {
	return gitstate.Inspect(ctx, cwd)
}

func currentStatus(ctx context.Context, cwd string) (statusOutput, error) {
	g, err := inspectGit(ctx, cwd)
	if err != nil {
		return statusOutput{}, err
	}
	root := workspace.Root(ctx, cwd)
	p := policy.Check(workspace.PolicyPath(root), g)
	receipts, err := receipt.List(workspace.ReceiptsDir(root))
	if err != nil {
		return statusOutput{}, err
	}
	checks, err := checkrecord.LoadAll(workspace.ChecksDir(root))
	if err != nil {
		return statusOutput{}, err
	}
	state := model.OK
	if p.State == model.Fail {
		state = model.Fail
	} else if !g.Clean || p.State == model.Warn {
		state = model.Warn
	}
	return statusOutput{KHZVersion: Version, Root: root, State: state, Git: g, Policy: p, ReceiptCount: len(receipts), CheckCount: len(checks)}, nil
}

func commandStatus(ctx context.Context, args []string, opts globalOptions, stdout, stderr io.Writer) int {
	jsonMode, rest := jsonFlag(args)
	if len(rest) != 0 {
		return usageError(stderr, "status accepts only --json")
	}
	cwd := workspace.Cwd()
	s, err := currentStatus(ctx, cwd)
	if err != nil {
		fmt.Fprintf(stderr, "khz: status: %v\n", err)
		return 1
	}
	if jsonMode {
		if err := writeJSON(stdout, s); err != nil {
			fmt.Fprintf(stderr, "khz: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, ui.Text(opts.Lang, "status_title"))
	allowColor := true
	if s.Policy.Policy != nil {
		allowColor = s.Policy.Policy.AllowColor
	}
	color := ui.ColorEnabled(stdout, opts.NoColor, allowColor)
	if !s.Git.InsideRepo {
		ui.Line(stdout, model.Info, "source", ui.Text(opts.Lang, "not_repo"), color)
	} else if s.Git.Clean {
		ui.Line(stdout, model.OK, "source", ui.Text(opts.Lang, "clean"), color)
	} else {
		ui.Line(stdout, model.Warn, "source", fmt.Sprintf("%s; staged=%d modified=%d untracked=%d", ui.Text(opts.Lang, "dirty"), s.Git.Staged, s.Git.Modified, s.Git.Untracked), color)
	}
	ui.Line(stdout, s.Policy.State, "policy", policyDetail(s.Policy, opts.Lang), color)
	ui.Line(stdout, model.Info, "receipts", fmt.Sprintf("%d", s.ReceiptCount), color)
	ui.Line(stdout, model.Info, "checks", fmt.Sprintf("%d", s.CheckCount), color)
	return 0
}

func policyDetail(p policy.CheckResult, lang string) string {
	if p.Valid {
		return ui.Text(lang, "policy_valid")
	}
	if len(p.Events) > 0 {
		return p.Events[0].Detail
	}
	return ui.Text(lang, "policy_missing")
}

func splitAfterDoubleDash(args []string) ([]string, error) {
	for i, a := range args {
		if a == "--" {
			if i != 0 {
				return nil, fmt.Errorf("unexpected arguments before --")
			}
			if len(args) == 1 {
				return nil, fmt.Errorf("missing command after --")
			}
			return args[1:], nil
		}
	}
	return nil, fmt.Errorf("expected -- before child command")
}

func commandRun(ctx context.Context, args []string, opts globalOptions, stdin io.Reader, stdout, stderr io.Writer, operation, checkName string) int {
	child, err := splitAfterDoubleDash(args)
	if err != nil {
		return usageError(stderr, "run: "+err.Error())
	}
	cwd := workspace.Cwd()
	root := workspace.Root(ctx, cwd)
	before, beforeErr := inspectGit(ctx, cwd)
	var beforePtr *gitstate.Snapshot
	if beforeErr == nil {
		beforePtr = &before
	}
	pcheck := policy.Check(workspace.PolicyPath(root), before)
	result := runner.Execute(ctx, cwd, child, stdin, stdout, stderr)
	after, afterErr := inspectGit(ctx, cwd)
	var afterPtr *gitstate.Snapshot
	if afterErr == nil {
		afterPtr = &after
	}
	id, err := receipt.NewID(result.StartedAt)
	if err != nil {
		fmt.Fprintf(stderr, "khz: receipt id: %v\n", err)
		return 1
	}
	r := receipt.Receipt{
		SchemaVersion: receipt.SchemaVersion, ReceiptID: id, KHZVersion: Version,
		StartedAt: result.StartedAt.Format(time.RFC3339Nano), FinishedAt: result.FinishedAt.Format(time.RFC3339Nano), DurationMS: result.DurationMS,
		Cwd: cwd, Operation: operation, Command: result.Executable, ArgsRedacted: result.ArgsRedacted, ExitCode: result.ExitCode, Termination: result.Termination,
		Status: result.Status, GitBefore: beforePtr, GitAfter: afterPtr, PolicyEvents: pcheck.Events, SideEffectObservability: "partial",
	}
	if checkName != "" {
		r.Checks = []receipt.CheckEvidence{{Name: checkName, Status: result.Status, ExitCode: result.ExitCode, DurationMS: result.DurationMS}}
	}
	path, err := receipt.Write(workspace.ReceiptsDir(root), r)
	if err != nil {
		fmt.Fprintf(stderr, "khz: write receipt: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "KHZ %s %s\n", result.Status, ui.Sanitize(result.Executable))
	fmt.Fprintf(stderr, "%s %s\n", ui.Text(opts.Lang, "receipt"), ui.Sanitize(path))
	if result.StartError {
		return 127
	}
	if result.ExitCode < 0 {
		return 1
	}
	return result.ExitCode
}

func commandCheck(ctx context.Context, args []string, opts globalOptions, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 4 || args[0] != "--name" {
		return usageError(stderr, "check usage: khz check --name <name> -- <command> [args...]")
	}
	name := args[1]
	if strings.TrimSpace(name) == "" {
		return usageError(stderr, "check name cannot be empty")
	}
	if args[2] != "--" {
		return usageError(stderr, "check requires -- before child command")
	}
	child := args[3:]
	if len(child) == 0 {
		return usageError(stderr, "check requires a child command")
	}
	cwd := workspace.Cwd()
	root := workspace.Root(ctx, cwd)
	before, beforeErr := inspectGit(ctx, cwd)
	var beforePtr *gitstate.Snapshot
	if beforeErr == nil {
		beforePtr = &before
	}
	pcheck := policy.Check(workspace.PolicyPath(root), before)
	res := runner.Execute(ctx, cwd, child, stdin, stdout, stderr)
	after, afterErr := inspectGit(ctx, cwd)
	var afterPtr *gitstate.Snapshot
	if afterErr == nil {
		afterPtr = &after
	}
	id, err := receipt.NewID(res.StartedAt)
	if err != nil {
		fmt.Fprintf(stderr, "khz: receipt id: %v\n", err)
		return 1
	}
	r := receipt.Receipt{SchemaVersion: receipt.SchemaVersion, ReceiptID: id, KHZVersion: Version, StartedAt: res.StartedAt.Format(time.RFC3339Nano), FinishedAt: res.FinishedAt.Format(time.RFC3339Nano), DurationMS: res.DurationMS, Cwd: cwd, Operation: "check", Command: res.Executable, ArgsRedacted: res.ArgsRedacted, ExitCode: res.ExitCode, Termination: res.Termination, Status: res.Status, GitBefore: beforePtr, GitAfter: afterPtr, Checks: []receipt.CheckEvidence{{Name: name, Status: res.Status, ExitCode: res.ExitCode, DurationMS: res.DurationMS}}, PolicyEvents: pcheck.Events, SideEffectObservability: "partial"}
	path, err := receipt.Write(workspace.ReceiptsDir(root), r)
	if err != nil {
		fmt.Fprintf(stderr, "khz: write receipt: %v\n", err)
		return 1
	}
	cr := checkrecord.Record{SchemaVersion: checkrecord.SchemaVersion, Name: name, Status: res.Status, Command: res.Executable, ArgsRedacted: res.ArgsRedacted, Cwd: cwd, StartedAt: res.StartedAt.Format(time.RFC3339Nano), FinishedAt: res.FinishedAt.Format(time.RFC3339Nano), DurationMS: res.DurationMS, ExitCode: res.ExitCode, ReceiptID: id}
	if _, err := checkrecord.Write(workspace.ChecksDir(root), cr); err != nil {
		fmt.Fprintf(stderr, "khz: write check state: %v\n", err)
		return 1
	}
	color := ui.ColorEnabled(stderr, opts.NoColor, pcheck.Policy == nil || pcheck.Policy.AllowColor)
	ui.Line(stderr, res.Status, name, fmt.Sprintf("exit=%d duration=%dms", res.ExitCode, res.DurationMS), color)
	fmt.Fprintf(stderr, "%s %s\n", ui.Text(opts.Lang, "receipt"), ui.Sanitize(path))
	if res.StartError {
		return 127
	}
	if res.ExitCode < 0 {
		return 1
	}
	return res.ExitCode
}

func commandBoard(ctx context.Context, args []string, opts globalOptions, stdout, stderr io.Writer) int {
	jsonMode, rest := jsonFlag(args)
	if len(rest) != 0 {
		return usageError(stderr, "board accepts only --json")
	}
	cwd := workspace.Cwd()
	root := workspace.Root(ctx, cwd)
	g, err := inspectGit(ctx, cwd)
	if err != nil {
		fmt.Fprintf(stderr, "khz: git: %v\n", err)
		return 1
	}
	p := policy.Check(workspace.PolicyPath(root), g)
	checks, err := checkrecord.LoadAll(workspace.ChecksDir(root))
	if err != nil {
		fmt.Fprintf(stderr, "khz: checks: %v\n", err)
		return 1
	}
	ids, err := receipt.List(workspace.ReceiptsDir(root))
	if err != nil {
		fmt.Fprintf(stderr, "khz: receipts: %v\n", err)
		return 1
	}
	latest := ""
	if len(ids) > 0 {
		latest = ids[0]
	}
	b := board.Build(g, p, checks, latest)
	if jsonMode {
		if err := writeJSON(stdout, b); err != nil {
			fmt.Fprintf(stderr, "khz: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, ui.Text(opts.Lang, "board_title"))
	allowColor := p.Policy == nil || p.Policy.AllowColor
	color := ui.ColorEnabled(stdout, opts.NoColor, allowColor)
	for _, e := range b.Entries {
		ui.Line(stdout, e.State, e.Name, e.Detail, color)
	}
	fmt.Fprintf(stdout, "\n%s  %s\n", ui.Text(opts.Lang, "decision"), b.Decision)
	if b.LatestReceipt != "" {
		fmt.Fprintf(stdout, "%s   %s\n", ui.Text(opts.Lang, "receipt"), ui.Sanitize(filepath.Join(workspace.ReceiptsDir(root), b.LatestReceipt+".json")))
	}
	return 0
}

func commandReceipt(ctx context.Context, args []string, opts globalOptions, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, "receipt requires list, show, or verify")
	}
	root := workspace.Root(ctx, workspace.Cwd())
	dir := workspace.ReceiptsDir(root)
	switch args[0] {
	case "list":
		jsonMode, rest := jsonFlag(args[1:])
		if len(rest) != 0 {
			return usageError(stderr, "receipt list accepts only --json")
		}
		ids, err := receipt.List(dir)
		if err != nil {
			fmt.Fprintf(stderr, "khz: receipt list: %v\n", err)
			return 1
		}
		if jsonMode {
			if err := writeJSON(stdout, ids); err != nil {
				return 1
			}
			return 0
		}
		for _, id := range ids {
			fmt.Fprintln(stdout, ui.Sanitize(id))
		}
		return 0
	case "show":
		if len(args) != 2 {
			return usageError(stderr, "receipt show requires exactly one id")
		}
		path, err := receipt.PathForID(dir, args[1])
		if err != nil {
			fmt.Fprintf(stderr, "khz: %v\n", err)
			return 2
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "khz: %v\n", err)
			return 1
		}
		_, _ = stdout.Write(data)
		return 0
	case "verify":
		jsonMode, rest := jsonFlag(args[1:])
		if len(rest) != 1 {
			return usageError(stderr, "receipt verify requires one id and optional --json")
		}
		path, err := receipt.PathForID(dir, rest[0])
		if err != nil {
			fmt.Fprintf(stderr, "khz: %v\n", err)
			return 2
		}
		v := receipt.VerifyFile(path)
		if jsonMode {
			_ = writeJSON(stdout, v)
		} else if v.Valid {
			fmt.Fprintf(stdout, "OK receipt %s %s\n", ui.Sanitize(v.ReceiptID), ui.Text(opts.Lang, "verified"))
		} else {
			fmt.Fprintf(stderr, "FAIL receipt %s: %s\n", ui.Sanitize(rest[0]), ui.Sanitize(v.Error))
		}
		if v.Valid {
			return 0
		}
		return 1
	default:
		return usageError(stderr, "receipt requires list, show, or verify")
	}
}

func commandPolicy(ctx context.Context, args []string, opts globalOptions, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, "policy requires init or check")
	}
	cwd := workspace.Cwd()
	root := workspace.Root(ctx, cwd)
	path := workspace.PolicyPath(root)
	switch args[0] {
	case "init":
		if len(args) != 1 {
			return usageError(stderr, "policy init takes no arguments")
		}
		if err := policy.Init(path); err != nil {
			if errors.Is(err, atomicfile.ErrExists) {
				fmt.Fprintln(stderr, "khz: policy already exists; refusing to overwrite")
				return 1
			}
			fmt.Fprintf(stderr, "khz: policy init: %v\n", err)
			return 1
		}
		ignorePath := filepath.Join(workspace.StateDir(root), ".gitignore")
		if _, err := os.Stat(ignorePath); errors.Is(err, os.ErrNotExist) {
			_ = atomicfile.WriteNew(ignorePath, []byte("receipts/\n"), 0o644)
		}
		configPath := workspace.ConfigPath(root)
		if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
			_ = atomicfile.WriteNew(configPath, []byte("{\n  \"schema_version\": 1\n}\n"), 0o600)
		}
		fmt.Fprintf(stdout, "policy initialized: %s\n", ui.Sanitize(path))
		return 0
	case "check":
		jsonMode, rest := jsonFlag(args[1:])
		if len(rest) != 0 {
			return usageError(stderr, "policy check accepts only --json")
		}
		g, err := inspectGit(ctx, cwd)
		if err != nil {
			fmt.Fprintf(stderr, "khz: git: %v\n", err)
			return 1
		}
		res := policy.Check(path, g)
		if jsonMode {
			_ = writeJSON(stdout, res)
		} else {
			color := ui.ColorEnabled(stdout, opts.NoColor, res.Policy == nil || res.Policy.AllowColor)
			for _, e := range res.Events {
				ui.Line(stdout, e.State, e.Kind, e.Detail, color)
			}
		}
		if res.Valid {
			return 0
		}
		return 1
	default:
		return usageError(stderr, "policy requires init or check")
	}
}

func commandGit(ctx context.Context, args []string, opts globalOptions, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "status" {
		return usageError(stderr, "git supports only: khz git status [--json]")
	}
	jsonMode, rest := jsonFlag(args[1:])
	if len(rest) != 0 {
		return usageError(stderr, "git status accepts only --json")
	}
	g, err := inspectGit(ctx, workspace.Cwd())
	if err != nil {
		fmt.Fprintf(stderr, "khz: git status: %v\n", err)
		return 1
	}
	if jsonMode {
		_ = writeJSON(stdout, g)
		return 0
	}
	color := ui.ColorEnabled(stdout, opts.NoColor, true)
	if !g.InsideRepo {
		ui.Line(stdout, model.Info, "git", ui.Text(opts.Lang, "not_repo"), color)
		return 0
	}
	state := model.OK
	detail := "clean"
	if !g.Clean {
		state = model.Warn
		detail = fmt.Sprintf("staged=%d modified=%d untracked=%d", g.Staged, g.Modified, g.Untracked)
	}
	branch := g.Branch
	if g.Detached {
		branch = "detached"
	}
	if g.Unborn {
		branch = "unborn:" + g.Branch
	}
	ui.Line(stdout, state, "git", branch+" "+detail, color)
	if g.Upstream != "" {
		ui.Line(stdout, model.Info, "upstream", fmt.Sprintf("%s ahead=%d behind=%d", g.Upstream, g.Ahead, g.Behind), color)
	}
	return 0
}
