package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"cahut.cloud/internal/doctor"
	"cahut.cloud/internal/update"
	"cahut.cloud/internal/version"
)

func runDoctor(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("doctor", stderr)
	data := dataFlag(fs)
	bin := fs.String("bin", update.DefaultBinPath, "the binary that the service runs")
	state := fs.String("state", update.DefaultStateDir, "directory of the state of updates")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: sudo cahut doctor [--data DIR]")
		fs.PrintDefaults()
	}
	if code, ok := parse(fs, args); !ok {
		return code
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	running, err := installedVersion(ctx, *bin)
	if err != nil {
		running = version.Get().Version
	}
	results := doctor.Run(ctx, doctor.Machine{}, doctor.Config{
		BinPath:   *bin,
		DataDir:   *data,
		StateDir:  *state,
		HealthURL: update.HealthURL,
		Version:   running,
	})
	return report(stdout, results)
}

// report prints the results of cahut doctor, one per line, and returns 1
// when one of them failed.
func report(w io.Writer, results []doctor.Result) int {
	width := 0
	for _, r := range results {
		width = max(width, len(r.Topic))
	}
	counts := map[doctor.Status]int{}
	for _, r := range results {
		counts[r.Status]++
		fmt.Fprintf(w, "%-5s %-*s  %s\n", r.Status, width, r.Topic, r.Text)
	}
	fmt.Fprintln(w)
	switch fails, warns := counts[doctor.Fail], counts[doctor.Warn]; {
	case fails > 0:
		found := count(fails, "problem", "problems")
		if warns > 0 {
			found += " and " + count(warns, "warning", "warnings")
		}
		fmt.Fprintf(w, "%s: fix what fails above, then run cahut doctor again.\n", found)
		return exitError
	case warns > 0:
		fmt.Fprintf(w, "No problem, %s.\n", count(warns, "warning", "warnings"))
	default:
		fmt.Fprintln(w, "No problem.")
	}
	return exitOK
}

func count(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}
