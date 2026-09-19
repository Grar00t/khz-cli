package cli

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Grar00t/khz-cli/internal/board"
	"github.com/Grar00t/khz-cli/internal/checkstore"
	"github.com/Grar00t/khz-cli/internal/execx"
	"github.com/Grar00t/khz-cli/internal/exitcode"
	"github.com/Grar00t/khz-cli/internal/gitx"
	"github.com/Grar00t/khz-cli/internal/i18n"
	"github.com/Grar00t/khz-cli/internal/policy"
	"github.com/Grar00t/khz-cli/internal/receipt"
	"github.com/Grar00t/khz-cli/internal/term"
	"github.com/Grar00t/khz-cli/internal/version"
)

func (a *app) cmdVersion(rest []string) int {
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("version"))
		return exitcode.OK
	}
	if len(rest) > 0 {
		a.errf("khz version: unexpected arguments")
		return exitcode.Usage
	}
	info := map[string]string{
		"name":    "khz",
		"version": version.Version,
		"go":      runtime.Version(),
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}
	if jsonFlag || a.g.JSON {
		return a.writeJSON(info)
	}
	a.outf("khz %s\n%s %s/%s\n", version.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return exitcode.OK
}

type doctorItem struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

func (a *app) cmdDoctor(rest []string) int {
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("doctor"))
		return exitcode.OK
	}
	_ = rest
	items := []doctorItem{
		{Name: "version", Status: term.INFO, Detail: "khz " + version.Version},
		{Name: "platform", Status: term.INFO, Detail: runtime.GOOS + "/" + runtime.GOARCH},
	}
	if _, err := exec.LookPath("git"); err != nil {
		items = append(items, doctorItem{Name: "git", Status: term.FAIL, Detail: "git executable not found"})
	} else {
		items = append(items, doctorItem{Name: "git", Status: term.OK, Detail: "git on PATH"})
	}
	if err := a.root.Ensure(); err != nil {
		items = append(items, doctorItem{Name: "state", Status: term.FAIL, Detail: err.Error()})
	} else {
		items = append(items, doctorItem{Name: "state", Status: term.OK, Detail: a.root.Dir})
	}
	colorDetail := "disabled"
	if a.color {
		colorDetail = "enabled"
	}
	items = append(items, doctorItem{Name: "color", Status: term.INFO, Detail: colorDetail})
	items = append(items, doctorItem{
		Name:   "observability",
		Status: term.INFO,
		Detail: "arbitrary commands record side_effect_observability=partial; KHZ does not prove absence of writes, network, or subprocesses",
	})
	fail := false
	for _, it := range items {
		if it.Status == term.FAIL {
			fail = true
		}
	}
	out := map[string]any{"status": term.OK, "items": items}
	if fail {
		out["status"] = term.FAIL
	}
	if jsonFlag || a.g.JSON {
		code := a.writeJSON(out)
		if fail {
			return exitcode.Fail
		}
		return code
	}
	for _, it := range items {
		a.outf("%s    %-16s %s\n", term.FormatState(it.Status, a.color), it.Name, it.Detail)
	}
	if fail {
		return exitcode.Fail
	}
	return exitcode.OK
}

func (a *app) loadPolicyAllowColor() {
	d, err := policy.Load(a.root.PolicyPath())
	if err != nil {
		return
	}
	if !d.AllowColor {
		a.color = false
	}
	if a.g.Lang == "en" && d.Language != "" {
		// command-line --lang wins; otherwise policy language
		if !langFlagPresent(a.env.Args) {
			a.g.Lang = d.Language
		}
	}
}

func langFlagPresent(args []string) bool {
	for _, a := range args {
		if a == "--lang" || strings.HasPrefix(a, "--lang=") {
			return true
		}
	}
	return false
}

func (a *app) cmdStatus(rest []string) int {
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("status"))
		return exitcode.OK
	}
	if len(rest) > 0 {
		a.errf("khz status: unexpected arguments")
		return exitcode.Usage
	}
	a.loadPolicyAllowColor()
	snap := gitx.Probe(a.ctx, a.env.CWD)
	checks, _ := checkstore.List(a.root.ChecksDir())
	pol := map[string]any{"present": false}
	if _, err := os.Stat(a.root.PolicyPath()); err == nil {
		res := policy.Check(a.root.PolicyPath(), snap)
		pol = map[string]any{
			"present":          true,
			"status":           res.Status,
			"protected_branch": res.ProtectedBranch,
			"valid":            res.Valid,
		}
	}
	payload := map[string]any{
		"khz_version": version.Version,
		"cwd":         a.env.CWD,
		"khz_root":    a.root.Dir,
		"git":         snap,
		"policy":      pol,
		"checks":      checks,
	}
	if jsonFlag || a.g.JSON {
		code := a.writeJSON(payload)
		if !snap.InRepo {
			return exitcode.Fail
		}
		return code
	}
	a.printGitHuman(snap)
	if len(checks) == 0 {
		a.outf("%s    %-10s %s\n", term.FormatState(term.SKIP, a.color), "checks", i18n.T(a.g.Lang, "none"))
	} else {
		for _, c := range checks {
			a.outf("%s    %-10s %s exit=%d\n", term.FormatState(c.Status, a.color), term.Sanitize(c.Name), term.Sanitize(c.Command), c.ExitCode)
		}
	}
	if !snap.InRepo {
		return exitcode.Fail
	}
	return exitcode.OK
}

func (a *app) printGitHuman(snap gitx.Snapshot) {
	st, detail := gitRow(snap, a.g.Lang)
	head := snap.Head
	if snap.Detached {
		head = "(detached)"
	}
	if snap.Initial {
		head = head + " (initial)"
	}
	a.outf("%s    %-10s %s\n", term.FormatState(st, a.color), "git", detail)
	if snap.InRepo {
		a.outf("%s    %-10s %s\n", term.FormatState(term.INFO, a.color), "head", term.Sanitize(head))
		if snap.ShortOID != "" {
			a.outf("%s    %-10s %s\n", term.FormatState(term.INFO, a.color), "oid", snap.ShortOID)
		}
		if snap.Ahead != nil || snap.Behind != nil {
			ahead, behind := 0, 0
			if snap.Ahead != nil {
				ahead = *snap.Ahead
			}
			if snap.Behind != nil {
				behind = *snap.Behind
			}
			a.outf("%s    %-10s +%d -%d %s\n", term.FormatState(term.INFO, a.color), "upstream", ahead, behind, term.Sanitize(snap.Upstream))
		}
	}
}

func gitRow(snap gitx.Snapshot, lang string) (string, string) {
	if !snap.InRepo {
		return term.SKIP, i18n.T(lang, "not_repo")
	}
	if snap.Initial {
		return term.INFO, i18n.T(lang, "unborn")
	}
	if snap.Clean {
		return term.OK, i18n.T(lang, "clean")
	}
	n := len(snap.Staged) + len(snap.Modified) + len(snap.Untracked) + len(snap.Unmerged) + len(snap.Renames)
	return term.WARN, fmt.Sprintf("%s (%d paths)", i18n.T(lang, "dirty"), n)
}

func (a *app) cmdGit(rest []string) int {
	if len(rest) == 0 {
		a.errf("khz git: expected status")
		return exitcode.Usage
	}
	sub := rest[0]
	rest = rest[1:]
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help || sub == "help" {
		a.outf("%s", commandHelp("git"))
		return exitcode.OK
	}
	if sub != "status" {
		a.errf("khz git: unknown subcommand %s", sub)
		return exitcode.Usage
	}
	if len(rest) > 0 {
		a.errf("khz git status: unexpected arguments")
		return exitcode.Usage
	}
	snap := gitx.Probe(a.ctx, a.env.CWD)
	if jsonFlag || a.g.JSON {
		_ = a.writeJSON(snap)
		if !snap.InRepo {
			return exitcode.Fail
		}
		return exitcode.OK
	}
	a.printGitHuman(snap)
	if !snap.InRepo {
		return exitcode.Fail
	}
	return exitcode.OK
}

func (a *app) cmdRun(rest []string) int {
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("run"))
		return exitcode.OK
	}
	flags, argv := splitDashDash(rest)
	if len(flags) > 0 {
		a.errf("khz run: unknown arguments %v (use: khz run -- <command> [args...])", flags)
		return exitcode.Usage
	}
	if len(argv) == 0 {
		a.errf("khz run: missing command after --")
		return exitcode.Usage
	}
	return a.execute("run", "", argv, jsonFlag || a.g.JSON)
}

func (a *app) cmdCheck(rest []string) int {
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("check"))
		return exitcode.OK
	}
	flags, argv := splitDashDash(rest)
	name, leftover, err := flagName(flags)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if len(leftover) > 0 {
		a.errf("khz check: unknown arguments %v", leftover)
		return exitcode.Usage
	}
	if name == "" {
		a.errf("khz check: --name is required")
		return exitcode.Usage
	}
	if len(argv) == 0 {
		a.errf("khz check: missing command after --")
		return exitcode.Usage
	}
	return a.execute("check", name, argv, jsonFlag || a.g.JSON)
}

func (a *app) execute(op, name string, argv []string, asJSON bool) int {
	if err := a.root.Ensure(); err != nil {
		a.errf("khz: %v", err)
		return exitcode.Fail
	}
	before := gitx.Probe(a.ctx, a.env.CWD)
	childOut, childErr := a.env.Stdout, a.env.Stderr
	if asJSON {
		childOut = a.env.Stderr
	}
	res := execx.Run(a.ctx, name, a.env.CWD, argv, childOut, childErr)
	after := gitx.Probe(a.ctx, a.env.CWD)
	id, err := receipt.NewID(a.now, a.env.Rand)
	if err != nil {
		id, err = receipt.NewID(a.now, rand.Read)
		if err != nil {
			a.errf("receipt id: %v", err)
			return exitcode.Fail
		}
	}
	rec := receipt.Receipt{
		SchemaVersion:           version.SchemaVersion,
		ReceiptID:               id,
		KHZVersion:              version.Version,
		StartedAt:               res.Started.UTC().Format(time.RFC3339Nano),
		FinishedAt:              res.Finished.UTC().Format(time.RFC3339Nano),
		DurationMS:              res.Finished.Sub(res.Started).Milliseconds(),
		CWD:                     a.env.CWD,
		Operation:               op,
		Command:                 res.Executable,
		ArgsRedacted:            res.ArgsRedacted,
		ExitCode:                res.ExitCode,
		Status:                  res.Status,
		GitBefore:               &before,
		GitAfter:                &after,
		Checks:                  []receipt.Check{},
		PolicyEvents:            []receipt.PolicyEvent{},
		Artifacts:               []string{},
		SideEffectObservability: receipt.SideEffectPartial,
	}
	if res.LookErr != "" {
		rec.Command = argv[0]
	}
	if op == "check" {
		ch := receipt.Check{
			Name:       name,
			Status:     res.Status,
			ExitCode:   res.ExitCode,
			DurationMS: rec.DurationMS,
			FinishedAt: rec.FinishedAt,
			Command:    rec.Command,
		}
		rec.Checks = []receipt.Check{ch}
		if err := checkstore.Save(a.root.ChecksDir(), ch); err != nil {
			a.errf("check store: %v", err)
		}
	}
	path, err := receipt.Write(a.root.ReceiptsDir(), &rec)
	if err != nil {
		a.errf("receipt write: %v", err)
		return exitcode.Fail
	}
	payload := map[string]any{
		"status":                    rec.Status,
		"exit_code":                 rec.ExitCode,
		"receipt_id":                rec.ReceiptID,
		"receipt":                   path,
		"side_effect_observability": rec.SideEffectObservability,
		"signaled":                  res.Signaled,
		"signal":                    res.Signal,
		"duration_ms":               rec.DurationMS,
		"command":                   rec.Command,
		"args_redacted":             rec.ArgsRedacted,
		"look_error":                res.LookErr,
	}
	if asJSON {
		code := a.writeJSON(payload)
		if rec.Status != term.OK {
			return exitcode.Fail
		}
		return code
	}
	label := op
	if name != "" {
		label = name
	}
	a.errf("%s    %-10s %s", term.FormatState(rec.Status, a.color), label, term.Sanitize(rec.Command))
	a.errf("%s    %-10s %s", term.FormatState(term.INFO, a.color), "receipt", path)
	if rec.Status != term.OK {
		return exitcode.Fail
	}
	return exitcode.OK
}

func (a *app) cmdBoard(rest []string) int {
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("board"))
		return exitcode.OK
	}
	if len(rest) > 0 {
		a.errf("khz board: unexpected arguments")
		return exitcode.Usage
	}
	a.loadPolicyAllowColor()
	snap := gitx.Probe(a.ctx, a.env.CWD)
	var rows []board.Row
	gst, gdetail := gitRow(snap, a.g.Lang)
	rows = append(rows, board.Row{State: gst, Name: "git", Detail: gdetail})
	if _, err := os.Stat(a.root.PolicyPath()); err == nil {
		pr := policy.Check(a.root.PolicyPath(), snap)
		detail := "valid"
		if !pr.Valid {
			detail = strings.Join(pr.Errors, "; ")
		} else if pr.ProtectedBranch {
			detail = i18n.T(a.g.Lang, "protected")
		}
		rows = append(rows, board.Row{State: pr.Status, Name: "policy", Detail: detail})
	} else {
		rows = append(rows, board.Row{State: term.SKIP, Name: "policy", Detail: "not initialized"})
	}
	checks, _ := checkstore.List(a.root.ChecksDir())
	if len(checks) == 0 {
		rows = append(rows, board.Row{State: term.SKIP, Name: "checks", Detail: i18n.T(a.g.Lang, "none")})
	} else {
		for _, c := range checks {
			rows = append(rows, board.Row{
				State:  c.Status,
				Name:   c.Name,
				Detail: fmt.Sprintf("%s exit=%d", c.Command, c.ExitCode),
			})
		}
	}
	ids, _ := receipt.List(a.root.ReceiptsDir())
	last := ""
	if len(ids) > 0 {
		last = a.root.ReceiptsDir() + string(os.PathSeparator) + ids[len(ids)-1] + ".json"
	}
	head := "no-repo"
	if snap.InRepo {
		ref := snap.Head
		if snap.Detached {
			ref = "detached"
		}
		if snap.Initial {
			ref = ref + "@unborn"
		} else if snap.ShortOID != "" {
			ref = ref + "@" + snap.ShortOID
		}
		head = ref
	}
	b := board.Board{
		Headline: head,
		Rows:     rows,
		Decision: board.Decide(rows),
		Receipt:  last,
	}
	if jsonFlag || a.g.JSON {
		return a.writeJSON(b)
	}
	width := term.Width(a.env.getenv, 80)
	a.outf("%s", board.Render(b, a.color, a.g.Lang, width))
	if b.Decision == "FAIL" {
		return exitcode.Fail
	}
	return exitcode.OK
}

func (a *app) cmdReceipt(rest []string) int {
	if len(rest) == 0 {
		a.errf("khz receipt: expected list|show|verify")
		return exitcode.Usage
	}
	sub := rest[0]
	rest = rest[1:]
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("receipt"))
		return exitcode.OK
	}
	switch sub {
	case "list":
		if len(rest) > 0 {
			a.errf("khz receipt list: unexpected arguments")
			return exitcode.Usage
		}
		ids, err := receipt.List(a.root.ReceiptsDir())
		if err != nil {
			a.errf("%v", err)
			return exitcode.Fail
		}
		if jsonFlag || a.g.JSON {
			return a.writeJSON(map[string]any{"receipts": ids})
		}
		for _, id := range ids {
			a.outf("%s\n", id)
		}
		return exitcode.OK
	case "show":
		if len(rest) != 1 {
			a.errf("khz receipt show <id>")
			return exitcode.Usage
		}
		path := receipt.Resolve(a.root.ReceiptsDir(), rest[0])
		b, err := os.ReadFile(path)
		if err != nil {
			a.errf("%v", err)
			return exitcode.Fail
		}
		a.outf("%s", string(b))
		return exitcode.OK
	case "verify":
		if len(rest) != 1 {
			a.errf("khz receipt verify <id>")
			return exitcode.Usage
		}
		path := receipt.Resolve(a.root.ReceiptsDir(), rest[0])
		raw, err := os.ReadFile(path)
		if err != nil {
			a.errf("%v", err)
			return exitcode.Fail
		}
		vr := receipt.VerifyBytes(raw)
		vr.Path = path
		if jsonFlag || a.g.JSON {
			_ = a.writeJSON(vr)
		} else {
			a.outf("%s    %-10s %s\n", term.FormatState(vr.Status, a.color), "verify", vr.Reason)
			if vr.ID != "" {
				a.outf("%s    %-10s %s\n", term.FormatState(term.INFO, a.color), "id", vr.ID)
			}
			if vr.OK {
				a.outf("%s    %-10s %s\n", term.FormatState(term.OK, a.color), "hash", vr.Hash)
			}
		}
		if !vr.OK {
			return exitcode.Fail
		}
		return exitcode.OK
	default:
		a.errf("khz receipt: unknown subcommand %s", sub)
		return exitcode.Usage
	}
}

func (a *app) cmdPolicy(rest []string) int {
	if len(rest) == 0 {
		a.errf("khz policy: expected init|check")
		return exitcode.Usage
	}
	sub := rest[0]
	rest = rest[1:]
	rest, jsonFlag, help, err := takeJSONHelp(rest)
	if err != nil {
		a.errf("%v", err)
		return exitcode.Usage
	}
	if help {
		a.outf("%s", commandHelp("policy"))
		return exitcode.OK
	}
	switch sub {
	case "init":
		if len(rest) > 0 {
			a.errf("khz policy init: unexpected arguments")
			return exitcode.Usage
		}
		if err := a.root.Ensure(); err != nil {
			a.errf("%v", err)
			return exitcode.Fail
		}
		d, err := policy.Init(a.root.PolicyPath())
		if err != nil {
			a.errf("%v", err)
			return exitcode.Fail
		}
		if jsonFlag || a.g.JSON {
			return a.writeJSON(map[string]any{"status": term.OK, "path": a.root.PolicyPath(), "policy": d})
		}
		a.outf("%s    %-10s %s\n", term.FormatState(term.OK, a.color), "policy", a.root.PolicyPath())
		a.outf("%s    %-10s %s\n", term.FormatState(term.INFO, a.color), "note", "local KHZ policy is not GitHub branch protection")
		return exitcode.OK
	case "check":
		if len(rest) > 0 {
			a.errf("khz policy check: unexpected arguments")
			return exitcode.Usage
		}
		snap := gitx.Probe(a.ctx, a.env.CWD)
		res := policy.Check(a.root.PolicyPath(), snap)
		if jsonFlag || a.g.JSON {
			_ = a.writeJSON(res)
			if !res.Valid {
				return exitcode.Fail
			}
			return exitcode.OK
		}
		a.outf("%s    %-10s valid=%v protected=%v\n", term.FormatState(res.Status, a.color), "policy", res.Valid, res.ProtectedBranch)
		for _, n := range res.Notes {
			a.outf("%s    %-10s %s\n", term.FormatState(term.INFO, a.color), "note", n)
		}
		for _, e := range res.Errors {
			a.errf("%s    %-10s %s", term.FormatState(term.FAIL, a.color), "error", e)
		}
		if !res.Valid {
			return exitcode.Fail
		}
		return exitcode.OK
	default:
		a.errf("khz policy: unknown subcommand %s", sub)
		return exitcode.Usage
	}
}
