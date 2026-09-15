//go:build windows

package exec

import (
	"os"
)

func isSafeHook(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return true
}
