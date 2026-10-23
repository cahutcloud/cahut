package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"cahut.cloud/internal/release"
	"cahut.cloud/internal/update"
	"cahut.cloud/internal/version"
)

// Settings files that install.sh writes. The updater reads them itself, so
// that sudo cahut update and the systemd unit see the same settings.
const (
	settingsFile = "/etc/cahut/cahut.env"
	updateFile   = "/etc/cahut/update.env" // readable by root only: the token to read the releases
)

func runUpdate(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("update", stderr)
	data := dataFlag(fs)
	var opts update.Options
	fs.BoolVar(&opts.Check, "check", false, "read the releases and say whether a newer one is out, without installing it")
	fs.StringVar(&opts.Version, "version", "", "install this release, such as 0.4.1, instead of the newest; never an older one")
	fs.StringVar(&opts.Binary, "binary", "", "install this cahut binary instead of a release, with the same check and rollback, without signature")
	fs.BoolVar(&opts.Unattended, "unattended", false, "run as systemd does: what the console asked for, or else what CAHUT_UPDATES says")
	bin := fs.String("bin", update.DefaultBinPath, "the binary that the service runs")
	state := fs.String("state", update.DefaultStateDir, "directory of the state of updates")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: sudo cahut update [--check] [--version VERSION] [--binary FILE]")
		fs.PrintDefaults()
	}
	if code, ok := parse(fs, args); !ok {
		return code
	}
	if os.Geteuid() != 0 {
		fmt.Fprintf(stderr, "cahut update: run it as root, with sudo %s update\n", update.DefaultBinPath)
		return exitError
	}
	settings := readSettings(settingsFile, updateFile)
	policy, err := update.ParsePolicy(settings("CAHUT_UPDATES"))
	if err != nil {
		fmt.Fprintf(stderr, "cahut update: %v\n", err)
		return exitError
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	running, err := installedVersion(ctx, *bin)
	if err != nil {
		fmt.Fprintf(stderr, "cahut update: %v\n", err)
		return exitError
	}
	repo := settings("CAHUT_REPO")
	if repo == "" {
		repo = "getcahut/cahut"
	}
	u := &update.Updater{
		Repo:      repo,
		API:       "https://api.github.com",
		Token:     settings("GITHUB_TOKEN"),
		Client:    &http.Client{Timeout: 10 * time.Minute},
		Keys:      release.Keys(),
		Arch:      runtime.GOARCH,
		Running:   running,
		Policy:    policy,
		BinPath:   *bin,
		DataDir:   *data,
		StateDir:  *state,
		Service:   update.Systemd{Unit: "cahut"},
		HealthURL: update.HealthURL,
		Wait:      time.Minute,
		Out:       stdout,
		Now:       time.Now,
	}
	st, err := u.Run(ctx, opts)
	if err != nil {
		fmt.Fprintf(stderr, "cahut update: %v\n", err)
		return exitError
	}
	if st.Last != nil && st.Last.Result == update.Failed {
		return exitError
	}
	return exitOK
}

// installedVersion returns the version of the binary that the service
// runs, which may not be the one running this command.
func installedVersion(ctx context.Context, bin string) (string, error) {
	out, err := exec.CommandContext(ctx, bin, "version").Output()
	if err != nil {
		if bin == update.DefaultBinPath {
			return "", fmt.Errorf("%s does not run, install Cahut with install.sh first: %w", bin, err)
		}
		return version.Get().Version, nil
	}
	fields := strings.Fields(string(out))
	if len(fields) < 2 {
		return "", fmt.Errorf("%s version prints %q", bin, out)
	}
	return fields[1], nil
}

// readSettings returns a function that reads a setting from the
// environment, or else from files, lines such as NAME=value.
func readSettings(files ...string) func(string) string {
	values := map[string]string{}
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		lines := bufio.NewScanner(f)
		for lines.Scan() {
			line := strings.TrimSpace(lines.Text())
			name, value, ok := strings.Cut(line, "=")
			if !ok || strings.HasPrefix(line, "#") {
				continue
			}
			values[strings.TrimSpace(name)] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
		f.Close()
	}
	return func(name string) string {
		if v, ok := os.LookupEnv(name); ok {
			return v
		}
		return values[name]
	}
}
