package cli

const rootHelp = `khz — structured decision and receipt layer for existing command-line workflows.

KHZ is not a terminal emulator.
KHZ is not a shell replacement.
KHZ is a structured decision and receipt layer for existing command-line workflows.

Usage:
  khz [global flags] <command> [flags] [-- <argv>...]

Global flags:
  --json          machine JSON on stdout where the command supports it
  --no-color      disable ANSI (also: NO_COLOR, TERM=dumb, non-TTY)
  --color         force ANSI even when stdout is not a TTY
  --lang en|ar    human-visible language (JSON keys stay English)
  -h, --help      help

Commands:
  version              print version
  doctor               inspect local toolchain and KHZ layout
  status               git + policy + last checks
  run -- <argv>        run a command, record a receipt (observability=partial)
  check --name N -- <argv>
                       named check, record receipt + board row
  board                human decision board from recorded state
  receipt list         list receipt ids
  receipt show <id>    print a receipt
  receipt verify <id>  accept untouched receipts; reject tamper/malformed/schema
  policy init          write .khz/policy.json (refuses overwrite)
  policy check         validate local policy (not GitHub branch protection)
  git status           parse git status --porcelain=v2 -z --branch

Exit codes:
  0  OK (WARN does not become a hidden PASS; it is printed as WARN)
  1  FAIL (check/run child failed, verify rejected, policy invalid, not a repo)
  2  usage error

Child processes are started with argv. KHZ does not use sh -c, cmd /c, or
Invoke-Expression. A receipt does not prove absence of filesystem, network,
registry, or subprocess side effects.
`

func commandHelp(cmd string) string {
	switch cmd {
	case "version":
		return "khz version [--json]\n\nPrint the KHZ version and Go runtime identity.\n"
	case "doctor":
		return "khz doctor [--json]\n\nInspect git availability, .khz writability, color policy, and observability limits.\n"
	case "status":
		return "khz status [--json]\n\nShow git porcelain-v2 state, policy presence, and last named checks.\n"
	case "run":
		return "khz run [--json] -- <command> [args...]\n\nRun argv directly. Record a receipt with side_effect_observability=partial.\n"
	case "check":
		return "khz check --name <name> [--json] -- <command> [args...]\n\nNamed check. exit 0 → OK, nonzero → FAIL. Stored for khz board.\n"
	case "board":
		return "khz board [--json]\n\nDecision board from real recorded checks and git state. Color is never the only signal.\n"
	case "receipt":
		return "khz receipt list\nkhz receipt show <id>\nkhz receipt verify <id>\n"
	case "policy":
		return "khz policy init\nkhz policy check [--json]\n\nLocal JSON policy only. Not GitHub branch protection.\n"
	case "git":
		return "khz git status [--json]\n\nUses git status --porcelain=v2 -z --branch. Does not parse human git status.\n"
	default:
		return rootHelp
	}
}
