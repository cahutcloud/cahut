package cli

import (
	"cmp"
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
	{name: "show", summary: "Print the kind and limits of a Cahut, with what it holds", run: runSettingsShow},
	{name: "set", summary: "Change the kind or limits of a Cahut; what you leave out stays as it is", run: runSettingsSet},
	{name: "links", summary: "Link the legal pages of a Cahut, or print its links; what you leave out stays as it is", run: runSettingsLinks},
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
	fmt.Fprintf(stdout, "Settings of %s:\n\n", at)
	return printSettings(stdout, about.Settings)
}

func runSettingsSet(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("settings set", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	kind := fs.String("kind", "", "collaborative or individual; the kind stays when left out")
	var people, admins, rooms, windows, perRoom, disk, roomsPP, windowsPP, diskPP limitFlag
	fs.Var(&people, "people", "the Cahut may have `N` people, administrators included, and disabled people do not count; none lifts the limit")
	fs.Var(&admins, "admins", "`N` of them may be administrators of the Cahut; none lifts the limit")
	fs.Var(&rooms, "rooms", "the Cahut may have `N` team Rooms; none lifts the limit")
	fs.Var(&windows, "windows", "the Cahut may hold `N` Windows out of the trash; none lifts the limit")
	fs.Var(&perRoom, "windows-per-room", "each Room may hold `N` Windows out of the trash; none lifts the limit")
	disk.bytes = true
	fs.Var(&disk, "disk", "the Cahut may keep `SIZE`, such as 10GB or 500MB, in units of 1000 bytes; none lifts the limit")
	fs.Var(&roomsPP, "rooms-per-person", "in an individual Cahut, each person may have `N` Rooms beside their personal Room; none lifts the limit")
	fs.Var(&windowsPP, "windows-per-person", "in an individual Cahut, each person may hold `N` Windows out of the trash; none lifts the limit")
	diskPP.bytes = true
	fs.Var(&diskPP, "disk-per-person", "in an individual Cahut, the Windows of each person may keep `SIZE`; none lifts the limit")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut settings set [--data DIR] [--domain DOMAIN] [--kind KIND] [--people N] [--admins N] [--rooms N] "+
			"[--windows N] [--windows-per-room N] [--disk SIZE] [--rooms-per-person N] [--windows-per-person N] [--disk-per-person SIZE]")
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
		if *kind == settings.Collaborative {
			// A collaborative Cahut has no limit per person: they go with
			// the kind, unless the flags set them, which the operation
			// then refuses.
			l.RoomsPerPerson, l.WindowsPerPerson, l.DiskPerPerson = nil, nil, nil
		}
		roomsPP.apply(&l.RoomsPerPerson)
		windowsPP.apply(&l.WindowsPerPerson)
		diskPP.apply(&l.DiskPerPerson)
		return operation.Do(ctx, c.Registry, settings.Set, admin, settings.SetCahutSettingsInput{Kind: *kind, Limits: l})
	})
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Set the settings of %s:\n\n", at)
	return printSettings(stdout, out)
}

// legalPage is a page that a Cahut links: the name of its flag, its name
// as the command line prints it and as its flag describes it, and its
// addresses.
type legalPage struct {
	flag, name, what string
	url, fr          *string
}

// legalPages are the pages that a Cahut links, with their addresses in l.
func legalPages(l *settings.LegalLinks) []legalPage {
	if l.FR == nil {
		l.FR = &settings.LegalPages{}
	}
	return []legalPage{
		{"legal-notice", "legal notice", "the legal notice", &l.LegalNotice, &l.FR.LegalNotice},
		{"terms", "terms", "the terms of use", &l.Terms, &l.FR.Terms},
		{"privacy", "privacy", "the privacy policy", &l.Privacy, &l.FR.Privacy},
		{"support", "support", "the page where people get help", &l.Support, &l.FR.Support},
	}
}

func runSettingsLinks(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("settings links", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var flags, frFlags [4]linkFlag
	for i, p := range legalPages(&settings.LegalLinks{}) {
		fs.Var(&flags[i], p.flag, "the `URL` of "+p.what+", which starts with https://; none removes the link, and its address in French")
		fs.Var(&frFlags[i], p.flag+"-fr", "the `URL` of "+p.what+" in French, where it differs, which the console links when it speaks French; none removes it")
	}
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut settings links [--data DIR] [--domain DOMAIN] [--legal-notice URL] [--terms URL] [--privacy URL] [--support URL] "+
			"[--legal-notice-fr URL] [--terms-fr URL] [--privacy-fr URL] [--support-fr URL]")
		fs.PrintDefaults()
	}
	if code, ok := parse(fs, args); !ok {
		return code
	}
	set := false
	for i := range flags {
		set = set || flags[i].set || frFlags[i].set
	}
	out, at, ok := withCahut(*data, *domain, stderr, func(ctx context.Context, c *cahut.Cahut) (settings.LegalLinks, error) {
		about, err := operation.Do(ctx, c.Registry, c.About, admin, cahut.GetCahutInput{})
		if err != nil || !set {
			return about.Legal, err
		}
		l := about.Legal
		for i, p := range legalPages(&l) {
			if flags[i].set {
				// A page that goes takes its address in French with it,
				// unless the flags set one, which the operation then
				// refuses.
				*p.url = flags[i].value
				if *p.url == "" {
					*p.fr = ""
				}
			}
			if frFlags[i].set {
				*p.fr = frFlags[i].value
			}
		}
		return operation.Do(ctx, c.Registry, settings.SetLegalLinks, admin, settings.SetLegalLinksInput{LegalLinks: l})
	})
	if !ok {
		return exitError
	}
	if set {
		fmt.Fprintf(stdout, "Set the legal pages of %s:\n\n", at)
	} else {
		fmt.Fprintf(stdout, "Legal pages of %s:\n\n", at)
	}
	return printLinks(stdout, out)
}

// linkFlag is a flag that changes the address of a legal page: a URL, or
// none to remove it.
type linkFlag struct {
	set   bool
	value string
}

func (f *linkFlag) String() string {
	if f == nil {
		return ""
	}
	return f.value
}

func (f *linkFlag) Set(s string) error {
	f.set = true
	f.value = s
	if s == "none" {
		f.value = ""
	}
	return nil
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

// printSettings prints the kind and limits of a Cahut, with what it holds.
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
	var u settings.Usage
	if s.Usage != nil {
		u = *s.Usage
	}
	fmt.Fprintf(w, "Kind: %s\n\n", s.Kind)
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "\tHOLDS\tLIMIT")
	fmt.Fprintf(tw, "people\t%d\t%s\n", u.People, limit(s.Limits.People, false))
	fmt.Fprintf(tw, "administrators\t%d\t%s\n", u.Admins, limit(s.Limits.Admins, false))
	fmt.Fprintf(tw, "team Rooms\t%d\t%s\n", u.Rooms, limit(s.Limits.Rooms, false))
	fmt.Fprintf(tw, "Windows\t%d\t%s\n", u.Windows, limit(s.Limits.Windows, false))
	fmt.Fprintf(tw, "Windows per Room\t\t%s\n", limit(s.Limits.WindowsPerRoom, false))
	fmt.Fprintf(tw, "disk\t%s\t%s\n", settings.Bytes(u.Disk), limit(s.Limits.Disk, true))
	if s.Kind == settings.Individual {
		fmt.Fprintf(tw, "Rooms per person\t\t%s\n", limit(s.Limits.RoomsPerPerson, false))
		fmt.Fprintf(tw, "Windows per person\t\t%s\n", limit(s.Limits.WindowsPerPerson, false))
		fmt.Fprintf(tw, "disk per person\t\t%s\n", limit(s.Limits.DiskPerPerson, true))
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	fmt.Fprint(w, "\nLegal pages:\n\n")
	return printLinks(w, s.Legal)
}

// printLinks prints the links to the legal pages of a Cahut, without the
// spaces that the column of the addresses in French leaves at the end of
// the lines without one.
func printLinks(w io.Writer, l settings.LegalLinks) int {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "\tADDRESS\tIN FRENCH")
	for _, p := range legalPages(&l) {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", p.name, cmp.Or(*p.url, "none"), *p.fr)
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	for line := range strings.Lines(b.String()) {
		fmt.Fprintln(w, strings.TrimRight(line, " \n"))
	}
	return exitOK
}
