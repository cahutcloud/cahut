package windata

import (
	"errors"
	"fmt"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// LandlockABI returns the version of Landlock that the kernel offers, or 0
// when it offers none.
func LandlockABI() int {
	abi, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET, 0, 0, unix.LANDLOCK_CREATE_RULESET_VERSION)
	if errno != 0 {
		return 0
	}
	return int(abi)
}

// rulesetAttr is struct landlock_ruleset_attr, of which a kernel reads as
// much as its version of Landlock knows.
type rulesetAttr struct {
	fs, net, scoped uint64
}

// Scopes of Landlock 6: abstract Unix sockets and signals outside the
// process.
const (
	scopeAbstractUnixSocket = 0x1
	scopeSignal             = 0x2
)

// fsAccess returns the rights on files that the version abi of Landlock
// knows, which the ruleset of a worker handles all.
func fsAccess(abi int) uint64 {
	access := uint64(unix.LANDLOCK_ACCESS_FS_EXECUTE | unix.LANDLOCK_ACCESS_FS_WRITE_FILE | unix.LANDLOCK_ACCESS_FS_READ_FILE |
		unix.LANDLOCK_ACCESS_FS_READ_DIR | unix.LANDLOCK_ACCESS_FS_REMOVE_DIR | unix.LANDLOCK_ACCESS_FS_REMOVE_FILE |
		unix.LANDLOCK_ACCESS_FS_MAKE_CHAR | unix.LANDLOCK_ACCESS_FS_MAKE_DIR | unix.LANDLOCK_ACCESS_FS_MAKE_REG |
		unix.LANDLOCK_ACCESS_FS_MAKE_SOCK | unix.LANDLOCK_ACCESS_FS_MAKE_FIFO | unix.LANDLOCK_ACCESS_FS_MAKE_BLOCK |
		unix.LANDLOCK_ACCESS_FS_MAKE_SYM)
	if abi >= 2 {
		access |= unix.LANDLOCK_ACCESS_FS_REFER
	}
	if abi >= 3 {
		access |= unix.LANDLOCK_ACCESS_FS_TRUNCATE
	}
	if abi >= 5 {
		access |= unix.LANDLOCK_ACCESS_FS_IOCTL_DEV
	}
	return access
}

// confine gives the worker of the database in dir its limits, then, when
// the kernel has Landlock, closes to it every file outside dir, and the
// network from Landlock 4. It applies to every thread of the process, and
// to those it starts later.
func confine(dir string) error {
	// Beyond its limit, a file write fails rather than kills the worker.
	signal.Ignore(syscall.SIGXFSZ)
	for _, l := range []struct {
		resource int
		value    uint64
	}{
		{unix.RLIMIT_CORE, 0},
		{unix.RLIMIT_FSIZE, 2 * MaxSize},
		{unix.RLIMIT_NOFILE, 64},
		{unix.RLIMIT_AS, 4 << 30},
	} {
		if l.resource == unix.RLIMIT_AS && raceEnabled {
			continue
		}
		if err := unix.Setrlimit(l.resource, &unix.Rlimit{Cur: l.value, Max: l.value}); err != nil {
			return fmt.Errorf("set limit %d: %w", l.resource, err)
		}
	}
	debug.SetMemoryLimit(2 * heapLimit)

	abi := LandlockABI()
	if abi < 1 {
		return nil // cahut doctor says that the workers are not confined
	}
	attr := rulesetAttr{fs: fsAccess(abi)}
	size := unsafe.Sizeof(attr.fs)
	if abi >= 4 {
		attr.net = unix.LANDLOCK_ACCESS_NET_BIND_TCP | unix.LANDLOCK_ACCESS_NET_CONNECT_TCP
		size += unsafe.Sizeof(attr.net)
	}
	if abi >= 6 {
		attr.scoped = scopeAbstractUnixSocket | scopeSignal
		size += unsafe.Sizeof(attr.scoped)
	}
	ruleset, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET, uintptr(unsafe.Pointer(&attr)), size, 0)
	if errno != 0 {
		return fmt.Errorf("create a Landlock ruleset: %w", errno)
	}
	defer unix.Close(int(ruleset))

	folder, err := unix.Open(dir, unix.O_PATH|unix.O_CLOEXEC|unix.O_DIRECTORY, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w", dir, err)
	}
	defer unix.Close(folder)
	allowed := uint64(unix.LANDLOCK_ACCESS_FS_READ_FILE | unix.LANDLOCK_ACCESS_FS_WRITE_FILE | unix.LANDLOCK_ACCESS_FS_READ_DIR |
		unix.LANDLOCK_ACCESS_FS_MAKE_REG | unix.LANDLOCK_ACCESS_FS_REMOVE_FILE)
	if abi >= 3 {
		allowed |= unix.LANDLOCK_ACCESS_FS_TRUNCATE
	}
	rule := unix.LandlockPathBeneathAttr{Allowed_access: allowed, Parent_fd: int32(folder)} //nolint:gosec // a file descriptor fits
	if _, _, errno := unix.Syscall6(unix.SYS_LANDLOCK_ADD_RULE, ruleset, unix.LANDLOCK_RULE_PATH_BENEATH,
		uintptr(unsafe.Pointer(&rule)), 0, 0, 0); errno != 0 {
		return fmt.Errorf("add the folder of the database to the Landlock ruleset: %w", errno)
	}
	if _, _, errno := syscall.AllThreadsSyscall(syscall.SYS_PRCTL, unix.PR_SET_NO_NEW_PRIVS, 1, 0); errno != 0 {
		return fmt.Errorf("set no_new_privs: %w", errno)
	}
	if _, _, errno := syscall.AllThreadsSyscall(unix.SYS_LANDLOCK_RESTRICT_SELF, ruleset, 0, 0); errno != 0 {
		if errors.Is(errno, syscall.ENOTSUP) {
			return errors.New("restrict every thread: the binary must be built without cgo")
		}
		return fmt.Errorf("restrict the worker with Landlock: %w", errno)
	}
	return nil
}

// detach makes the worker of cmd die with the service.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
}
