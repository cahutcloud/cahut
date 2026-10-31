//go:build !linux

package windata

import (
	"fmt"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"syscall"
)

// restrictsThreads is false: nothing restricts a worker outside Linux.
const restrictsThreads = false

// LandlockABI returns 0: Landlock is a module of Linux.
func LandlockABI() int { return 0 }

// confine gives the worker its limits. Outside Linux, nothing closes the
// files of the machine to it: cahut doctor says so.
func confine(string) error {
	signal.Ignore(syscall.SIGXFSZ)
	for _, l := range []struct {
		resource int
		value    uint64
	}{
		{syscall.RLIMIT_CORE, 0},
		{syscall.RLIMIT_FSIZE, 2 * MaxSize},
	} {
		if err := syscall.Setrlimit(l.resource, &syscall.Rlimit{Cur: l.value, Max: l.value}); err != nil {
			return fmt.Errorf("set limit %d: %w", l.resource, err)
		}
	}
	debug.SetMemoryLimit(2 * heapLimit)
	return nil
}

// detach does nothing outside Linux: the worker ends with its standard
// input, which ends with the service.
func detach(*exec.Cmd) {}
