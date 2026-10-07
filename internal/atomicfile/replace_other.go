//go:build !windows

package atomicfile

import "os"

func platformReplace(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
