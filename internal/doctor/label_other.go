//go:build !linux

package doctor

import "errors"

// label reads the SELinux label of a file: there is none outside Linux.
func label(string) (string, error) { return "", errors.New("no SELinux on this system") }
