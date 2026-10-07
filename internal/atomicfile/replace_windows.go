//go:build windows

package atomicfile

import (
	"os"
	"syscall"
	"unsafe"
)

var replaceFileW = syscall.NewLazyDLL("kernel32.dll").NewProc("ReplaceFileW")

func platformReplace(oldPath, newPath string) error {
	if _, err := os.Stat(newPath); err != nil {
		if os.IsNotExist(err) {
			return os.Rename(oldPath, newPath)
		}
		return err
	}

	target, err := syscall.UTF16PtrFromString(newPath)
	if err != nil {
		return err
	}
	replacement, err := syscall.UTF16PtrFromString(oldPath)
	if err != nil {
		return err
	}

	r1, _, callErr := replaceFileW.Call(
		uintptr(unsafe.Pointer(target)),
		uintptr(unsafe.Pointer(replacement)),
		0,
		0,
		0,
		0,
	)
	if r1 == 0 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return callErr
		}
		return syscall.EINVAL
	}
	return nil
}
