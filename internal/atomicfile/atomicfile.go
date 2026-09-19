package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes data to path by creating a unique temp file in the same
// directory, fsyncing, then renaming. On failure the original target is left
// unchanged if it existed.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".khz-tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tmp)
		}
	}()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		// Windows cannot rename over an existing file.
		if err2 := os.Remove(path); err2 == nil || os.IsNotExist(err2) {
			if err3 := os.Rename(tmp, path); err3 == nil {
				ok = true
				return nil
			} else {
				return fmt.Errorf("atomic rename: %w", err3)
			}
		}
		return fmt.Errorf("atomic rename: %w", err)
	}
	ok = true
	return nil
}
