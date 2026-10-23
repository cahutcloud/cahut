//go:build !linux

package doctor

import (
	"errors"
	"syscall"
)

// label reads the SELinux label of a file: there is none outside Linux.
func label(string) (string, error) { return "", errors.New("no SELinux on this system") }

// freeBytes returns the space left to unprivileged users on the file
// system of path.
func freeBytes(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return st.Bavail * uint64(st.Bsize), nil
}
