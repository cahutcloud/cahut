package doctor

import (
	"strings"
	"syscall"
)

// label reads the SELinux label of a file, from its extended attribute.
func label(name string) (string, error) {
	buf := make([]byte, 256)
	n, err := syscall.Getxattr(name, "security.selinux", buf)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(buf[:n]), "\x00"), nil
}

// freeBytes returns the space left to unprivileged users on the file
// system of path.
func freeBytes(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return st.Bavail * uint64(st.Bsize), nil //nolint:gosec // a block size is positive
}
