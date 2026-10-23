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
