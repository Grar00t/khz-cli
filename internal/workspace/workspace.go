package workspace

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Grar00t/khz-cli/internal/gitstate"
)

func Root(ctx context.Context, cwd string) string {
	if snap, err := gitstate.Inspect(ctx, cwd); err == nil && snap.InsideRepo && snap.Root != "" {
		return snap.Root
	}
	abs, err := filepath.Abs(cwd)
	if err == nil {
		return abs
	}
	return cwd
}

func StateDir(root string) string    { return filepath.Join(root, ".khz") }
func ReceiptsDir(root string) string { return filepath.Join(StateDir(root), "receipts") }
func ChecksDir(root string) string   { return filepath.Join(StateDir(root), "checks") }
func PolicyPath(root string) string  { return filepath.Join(StateDir(root), "policy.json") }
func ConfigPath(root string) string  { return filepath.Join(StateDir(root), "config.json") }

func Cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
