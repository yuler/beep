//go:build !windows

package exec

import (
	"os"
	"syscall"
)

func isSafeHook(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if info.Mode().Perm()&0o022 != 0 {
		return false
	}
	if info.Mode().Perm()&0o111 == 0 {
		return false
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if int(stat.Uid) != os.Getuid() {
			return false
		}
	}
	return true
}
