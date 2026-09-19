package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("KHZ_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	idx := -1
	for i, a := range args {
		if a == "--" {
			idx = i
			break
		}
	}
	if idx >= 0 && idx+1 < len(args) {
		switch args[idx+1] {
		case "ok":
			fmt.Fprint(os.Stdout, strings.Join(args[idx+2:], "|"))
			os.Exit(0)
		case "fail":
			os.Exit(7)
		}
	}
	os.Exit(9)
}

func helperArgs(mode string, extras ...string) []string {
	args := []string{os.Args[0], "-test.run=TestHelperProcess", "--", mode}
	return append(args, extras...)
}

func TestExecuteExitZeroAndArgvSpaces(t *testing.T) {
	old := os.Getenv("KHZ_HELPER_PROCESS")
	_ = os.Setenv("KHZ_HELPER_PROCESS", "1")
	t.Cleanup(func() { _ = os.Setenv("KHZ_HELPER_PROCESS", old) })
	var out bytes.Buffer
	res := Execute(context.Background(), t.TempDir(), helperArgs("ok", "two words", "ثلاثة"), nil, &out, &out)
	if res.ExitCode != 0 || res.Status != "OK" {
		t.Fatalf("result: %+v", res)
	}
	if out.String() != "two words|ثلاثة" {
		t.Fatalf("argv corrupted: %q", out.String())
	}
}

func TestExecuteNonzero(t *testing.T) {
	old := os.Getenv("KHZ_HELPER_PROCESS")
	_ = os.Setenv("KHZ_HELPER_PROCESS", "1")
	t.Cleanup(func() { _ = os.Setenv("KHZ_HELPER_PROCESS", old) })
	res := Execute(context.Background(), t.TempDir(), helperArgs("fail"), nil, &bytes.Buffer{}, &bytes.Buffer{})
	if res.ExitCode != 7 || res.Status != "FAIL" {
		t.Fatalf("result: %+v", res)
	}
}

func TestExecuteMissingProgram(t *testing.T) {
	res := Execute(context.Background(), t.TempDir(), []string{"khz-program-that-does-not-exist-xyz"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if !res.StartError || res.ExitCode != -1 {
		t.Fatalf("result: %+v", res)
	}
}
