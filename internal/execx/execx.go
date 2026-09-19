package execx

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/Grar00t/khz-cli/internal/redact"
	"github.com/Grar00t/khz-cli/internal/term"
)

// Result is the recorded outcome of a child process started with argv (no shell).
type Result struct {
	Name         string
	Executable   string
	Args         []string
	ArgsRedacted []string
	CWD          string
	Started      time.Time
	Finished     time.Time
	ExitCode     int
	Signaled     bool
	Signal       string
	Status       string
	LookErr      string
}

// Run starts exe with argv (argv[0] is exe as passed; extra args follow).
func Run(ctx context.Context, name, cwd string, argv []string, stdout, stderr io.Writer) Result {
	if ctx == nil {
		ctx = context.Background()
	}
	res := Result{
		Name:         name,
		CWD:          cwd,
		Args:         append([]string(nil), argv...),
		ArgsRedacted: redact.Args(argv),
		Started:      time.Now().UTC(),
		Status:       term.FAIL,
		ExitCode:     1,
	}
	if len(argv) == 0 {
		res.LookErr = "empty command"
		res.Finished = time.Now().UTC()
		return res
	}
	exe := argv[0]
	resolved, err := exec.LookPath(exe)
	if err != nil {
		res.Executable = exe
		res.LookErr = err.Error()
		res.Finished = time.Now().UTC()
		return res
	}
	res.Executable = resolved
	cmd := exec.CommandContext(ctx, resolved, argv[1:]...)
	cmd.Dir = cwd
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	res.Finished = time.Now().UTC()
	if err == nil {
		res.ExitCode = 0
		res.Status = term.OK
		return res
	}
	if ee, ok := err.(*exec.ExitError); ok {
		res.ExitCode = ee.ExitCode()
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
			if ws.Signaled() {
				res.Signaled = true
				res.Signal = ws.Signal().String()
			}
		}
		return res
	}
	res.LookErr = err.Error()
	return res
}
