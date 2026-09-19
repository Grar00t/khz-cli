package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Grar00t/khz-cli/internal/exitcode"
	"github.com/Grar00t/khz-cli/internal/state"
	"github.com/Grar00t/khz-cli/internal/term"
)

// Env is the process environment. Tests inject buffers and a fake clock.
type Env struct {
	Args      []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	CWD       string
	Getenv    func(string) string
	StdoutTTY bool
	Now       func() time.Time
	Rand      func([]byte) (int, error)
}

type globals struct {
	JSON       bool
	NoColor    bool
	ForceColor bool
	Lang       string
	Help       bool
}

type app struct {
	env   Env
	g     globals
	color bool
	root  state.Root
	now   time.Time
	ctx   context.Context
}

func (e Env) getenv(k string) string {
	if e.Getenv != nil {
		return e.Getenv(k)
	}
	return os.Getenv(k)
}

func (e Env) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now().UTC()
}

func (a *app) errf(format string, args ...any) {
	fmt.Fprintf(a.env.Stderr, format, args...)
	if !strings.HasSuffix(format, "\n") {
		fmt.Fprintln(a.env.Stderr)
	}
}

func (a *app) outf(format string, args ...any) {
	fmt.Fprintf(a.env.Stdout, format, args...)
}

func (a *app) writeJSON(v any) int {
	enc := json.NewEncoder(a.env.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		a.errf("json encode: %v", err)
		return exitcode.Fail
	}
	return exitcode.OK
}

func splitArgs(args []string) (g globals, cmd string, rest []string, err error) {
	g.Lang = "en"
	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case a == "--":
			return g, "", nil, fmt.Errorf("missing command before --")
		case a == "--json":
			g.JSON = true
		case a == "--no-color":
			g.NoColor = true
		case a == "--color":
			g.ForceColor = true
		case a == "--help" || a == "-h":
			g.Help = true
		case a == "--lang":
			if i+1 >= len(args) {
				return g, "", nil, fmt.Errorf("--lang requires a value")
			}
			i++
			g.Lang = args[i]
		case strings.HasPrefix(a, "--lang="):
			g.Lang = strings.TrimPrefix(a, "--lang=")
		case strings.HasPrefix(a, "-"):
			return g, "", nil, fmt.Errorf("unknown flag %s", a)
		default:
			cmd = a
			rest = args[i+1:]
			return g, cmd, rest, nil
		}
		i++
	}
	return g, "", nil, nil
}

func takeJSONHelp(rest []string) (rest2 []string, jsonFlag, help bool, err error) {
	out := make([]string, 0, len(rest))
	for i := 0; i < len(rest); i++ {
		a := rest[i]
		if a == "--" {
			out = append(out, rest[i:]...)
			return out, jsonFlag, help, nil
		}
		switch a {
		case "--json":
			jsonFlag = true
		case "--help", "-h":
			help = true
		default:
			out = append(out, a)
		}
	}
	return out, jsonFlag, help, nil
}

func splitDashDash(rest []string) (flags, argv []string) {
	for i, a := range rest {
		if a == "--" {
			return rest[:i], rest[i+1:]
		}
	}
	return rest, nil
}

func flagName(flags []string) (string, []string, error) {
	name := ""
	var leftover []string
	for i := 0; i < len(flags); i++ {
		a := flags[i]
		switch {
		case a == "--name":
			if i+1 >= len(flags) {
				return "", nil, fmt.Errorf("--name requires a value")
			}
			i++
			name = flags[i]
		case strings.HasPrefix(a, "--name="):
			name = strings.TrimPrefix(a, "--name=")
		default:
			leftover = append(leftover, a)
		}
	}
	return name, leftover, nil
}

// Main is the testable entrypoint. args[0] is the program name.
func Main(env Env) int {
	if env.Stdin == nil {
		env.Stdin = os.Stdin
	}
	if env.Stdout == nil {
		env.Stdout = os.Stdout
	}
	if env.Stderr == nil {
		env.Stderr = os.Stderr
	}
	if env.CWD == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(env.Stderr, "cwd: %v\n", err)
			return exitcode.Fail
		}
		env.CWD = wd
	}

	args := env.Args
	if len(args) > 0 {
		args = args[1:]
	}
	g, cmd, rest, err := splitArgs(args)
	if err != nil {
		fmt.Fprintf(env.Stderr, "khz: %v\n", err)
		return exitcode.Usage
	}
	lang := strings.ToLower(g.Lang)
	if lang != "en" && lang != "ar" {
		fmt.Fprintf(env.Stderr, "khz: --lang must be en or ar\n")
		return exitcode.Usage
	}
	g.Lang = lang

	if g.Help && cmd == "" {
		fmt.Fprint(env.Stdout, rootHelp)
		return exitcode.OK
	}
	if cmd == "" {
		fmt.Fprint(env.Stderr, rootHelp)
		return exitcode.Usage
	}

	a := &app{
		env:  env,
		g:    g,
		root: state.Locate(env.CWD),
		now:  env.now(),
		ctx:  context.Background(),
	}
	cc := term.ColorConfig{
		NoColor:    g.NoColor,
		ForceColor: g.ForceColor,
		StdoutTTY:  env.StdoutTTY,
		Getenv:     env.getenv,
		AllowColor: true,
	}
	a.color = cc.Enabled()

	if g.Help {
		a.outf("%s", commandHelp(cmd))
		return exitcode.OK
	}

	switch cmd {
	case "version":
		return a.cmdVersion(rest)
	case "doctor":
		return a.cmdDoctor(rest)
	case "status":
		return a.cmdStatus(rest)
	case "run":
		return a.cmdRun(rest)
	case "check":
		return a.cmdCheck(rest)
	case "board":
		return a.cmdBoard(rest)
	case "receipt":
		return a.cmdReceipt(rest)
	case "policy":
		return a.cmdPolicy(rest)
	case "git":
		return a.cmdGit(rest)
	case "help":
		if len(rest) > 0 {
			a.outf("%s", commandHelp(rest[0]))
			return exitcode.OK
		}
		a.outf("%s", rootHelp)
		return exitcode.OK
	default:
		a.errf("khz: unknown command %s", cmd)
		return exitcode.Usage
	}
}

// RunOS is the process wrapper.
func RunOS() int {
	tty := false
	if st, err := os.Stdout.Stat(); err == nil {
		tty = st.Mode()&os.ModeCharDevice != 0
	}
	wd, _ := os.Getwd()
	return Main(Env{
		Args:      os.Args,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		CWD:       wd,
		Getenv:    os.Getenv,
		StdoutTTY: tty,
	})
}
