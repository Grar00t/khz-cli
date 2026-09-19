package runner

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/Grar00t/khz-cli/internal/model"
	"github.com/Grar00t/khz-cli/internal/redact"
)

type Result struct {
	Executable   string      `json:"executable"`
	ArgsRedacted []string    `json:"args_redacted"`
	Cwd          string      `json:"cwd"`
	StartedAt    time.Time   `json:"-"`
	FinishedAt   time.Time   `json:"-"`
	DurationMS   int64       `json:"duration_ms"`
	ExitCode     int         `json:"exit_code"`
	Termination  string      `json:"termination,omitempty"`
	Status       model.State `json:"status"`
	StartError   bool        `json:"start_error"`
}

func Execute(ctx context.Context, cwd string, argv []string, stdin io.Reader, stdout, stderr io.Writer) Result {
	started := time.Now().UTC()
	res := Result{Cwd: cwd, StartedAt: started, ExitCode: -1, Status: model.Fail}
	if len(argv) == 0 {
		res.Termination = "no executable provided"
		res.FinishedAt = time.Now().UTC()
		res.DurationMS = res.FinishedAt.Sub(started).Milliseconds()
		res.StartError = true
		return res
	}
	res.Executable = argv[0]
	res.ArgsRedacted = redact.Args(argv[1:])
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = cwd
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	finished := time.Now().UTC()
	res.FinishedAt = finished
	res.DurationMS = finished.Sub(started).Milliseconds()
	if cmd.ProcessState != nil {
		res.ExitCode = cmd.ProcessState.ExitCode()
		res.Termination = cmd.ProcessState.String()
	}
	if err == nil {
		res.ExitCode = 0
		res.Status = model.OK
		res.Termination = "exit 0"
		return res
	}
	if cmd.ProcessState == nil {
		res.StartError = true
		res.Termination = fmt.Sprintf("start error: %v", err)
		return res
	}
	res.Status = model.Fail
	return res
}
