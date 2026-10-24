package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"

	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

var settingsCommands = []command{
	{name: "show", summary: "Print the limits of a Cahut, with what it holds", run: runSettingsShow},
	{name: "set", summary: "Change limits of a Cahut; those you leave out stay as they are", run: runSettingsSet},
}

func runSettingsShow(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("settings show", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	about, at, ok := withCahut(*data, *domain, stderr, func(ctx context.Context, c *cahut.Cahut) (cahut.About, error) {
		return operation.Do(ctx, c.Registry, c.About, admin, cahut.GetCahutInput{})
	})
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Limits of %s:\n\n", at)
	return printSettings(stdout, about.Settings)
}

func runSettingsSet(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("settings set", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var people, admins, rooms, windows, perRoom, disk limitFlag
	fs.Var(&people, "people", "the Cahut may have `N` people, administrators included, and disabled people do not count; none lifts the limit")
	fs.Var(&admins, "admins", "`N` of them may be administrators of the Cahut; none lifts the limit")
	fs.Var(&rooms, "rooms", "the Cahut may have `N` team Rooms; none lifts the limit")
	fs.Var(&windows, "windows", "the Cahut may hold `N` Windows out of the trash; none lifts the limit")
	fs.Var(&perRoom, "windows-per-room", "each Room may hold `N` Windows out of the trash; none lifts the limit")
	disk.bytes = true
	fs.Var(&disk, "disk", "the Cahut may keep `SIZE`, such as 10GB or 500MB, in units of 1000 bytes; none lifts the limit")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut settings set [--data DIR] [--domain DOMAIN] [--people N] [--admins N] [--rooms N] "+
			"[--windows N] [--windows-per-room N] [--disk SIZE]")
		fs.PrintDefaults()
	}
	if code, ok := parse(fs, args); !ok {
		return code
	}
	out, at, ok := withCahut(*data, *domain, stderr, func(ctx context.Context, c *cahut.Cahut) (settings.Settings, error) {
		about, err := operation.Do(ctx, c.Registry, c.About, admin, cahut.GetCahutInput{})
		if err != nil {
			return settings.Settings{}, err
		}
		l := about.Limits
		people.apply(&l.People)
		admins.apply(&l.Admins)
		rooms.apply(&l.Rooms)
		windows.apply(&l.Windows)
		perRoom.apply(&l.WindowsPerRoom)
		disk.apply(&l.Disk)
		return operation.Do(ctx, c.Registry, settings.Set, admin, settings.SetCahutSettingsInput{Limits: l})
	})
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Set the limits of %s:\n\n", at)
	return printSettings(stdout, out)
}

// limitFlag is a flag that changes one limit: a number, or none to lift it.
type limitFlag struct {
	set   bool
	value *int64
	bytes bool // the limit counts bytes, given as a size such as 10GB
}

func (f *limitFlag) String() string {
	if f == nil || f.value == nil {
		return ""
	}
	return strconv.FormatInt(*f.value, 10)
}

func (f *limitFlag) Set(s string) error {
	f.set = true
	if s == "none" {
		f.value = nil
		return nil
	}
	if f.bytes {
		n, err := parseSize(s)
		if err != nil {
			return err
		}
		f.value = &n
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return errors.New("give a number, or none")
	}
	f.value = &n
	return nil
}

// apply sets the limit l to the flag's value, when the flag was given.
func (f *limitFlag) apply(l **int64) {
	if f.set {
		*l = f.value
	}
}

// parseSize reads a size such as 10GB, 1.5 GB, 500MB or 2000000, in units
// of 1000 bytes, as hosts sell disk.
func parseSize(s string) (int64, error) {
	bad := fmt.Errorf("size %q: give a number with a unit, such as 10GB or 500MB, or none", s)
	t := strings.ToUpper(strings.TrimSpace(s))
	units := []struct {
		suffix string
		bytes  float64
	}{{"TB", 1e12}, {"GB", 1e9}, {"MB", 1e6}, {"KB", 1e3}, {"T", 1e12}, {"G", 1e9}, {"M", 1e6}, {"K", 1e3}, {"B", 1}}
	mult := 1.0
	for _, u := range units {
		if rest, ok := strings.CutSuffix(t, u.suffix); ok {
			t, mult = strings.TrimSpace(rest), u.bytes
			break
		}
	}
	v, err := strconv.ParseFloat(t, 64)
	if err != nil || v < 0 || math.IsInf(v, 0) || math.IsNaN(v) || v*mult >= math.MaxInt64 {
		return 0, bad
	}
	return int64(math.Round(v * mult)), nil
}

// printSettings prints the limits of a Cahut, with what it holds.
func printSettings(w io.Writer, s settings.Settings) int {
	limit := func(l *int64, bytes bool) string {
		switch {
		case l == nil:
			return "none"
		case bytes:
			return settings.Bytes(*l)
		default:
			return strconv.FormatInt(*l, 10)
		}
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "\tHOLDS\tLIMIT")
	fmt.Fprintf(tw, "people\t%d\t%s\n", s.Usage.People, limit(s.Limits.People, false))
	fmt.Fprintf(tw, "administrators\t%d\t%s\n", s.Usage.Admins, limit(s.Limits.Admins, false))
	fmt.Fprintf(tw, "team Rooms\t%d\t%s\n", s.Usage.Rooms, limit(s.Limits.Rooms, false))
	fmt.Fprintf(tw, "Windows\t%d\t%s\n", s.Usage.Windows, limit(s.Limits.Windows, false))
	fmt.Fprintf(tw, "Windows per Room\t\t%s\n", limit(s.Limits.WindowsPerRoom, false))
	fmt.Fprintf(tw, "disk\t%s\t%s\n", settings.Bytes(s.Usage.Disk), limit(s.Limits.Disk, true))
	if err := tw.Flush(); err != nil {
		return exitError
	}
	return exitOK
}
