// Package doctor diagnoses a server of Cahut: the system and its clock, the
// service, its data, its ports, the firewall and SELinux, the address and
// the certificate of each Cahut, the services it needs outside, and its
// updates. Each check says what it found and, when something is wrong,
// what to do. It never opens the databases: it asks the running service,
// on the machine itself. See docs/plan-lot-7.md.
package doctor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"cahut.cloud/internal/release"
	"cahut.cloud/internal/update"
)

// Status is how a check went.
type Status string

// Statuses of a check.
const (
	OK   Status = "ok"
	Warn Status = "warn" // works, but should be looked at
	Fail Status = "fail" // does not work
	Info Status = "info" // nothing to check, something to know
)

// Result is what one check found.
type Result struct {
	Status Status
	Topic  string
	Text   string
}

// System is what the checks read of the machine. Tests replace it.
type System interface {
	// Command runs a program and returns what it printed, trimmed.
	Command(ctx context.Context, name string, args ...string) (string, error)
	ReadFile(name string) ([]byte, error)
	Stat(name string) (fs.FileInfo, error)
	// Owner returns the user that owns a file, by name.
	Owner(name string) (string, error)
	// Label returns the SELinux label of a file.
	Label(name string) (string, error)
	FreeBytes(path string) (uint64, error)
	LocalAddrs() ([]net.IP, error)
	LookupIP(ctx context.Context, host string) ([]net.IP, error)
	// Dial opens a TCP connection to addr, and closes it.
	Dial(ctx context.Context, addr string) error
	// Get reads url, checking certificates as browsers do.
	Get(ctx context.Context, url string) (int, []byte, error)
	// Certificate returns the certificate that addr serves for name.
	Certificate(ctx context.Context, addr, name string) (Certificate, error)
	Root() bool
	Now() time.Time
}

// Certificate is a certificate that the service serves.
type Certificate struct {
	NotBefore, NotAfter time.Time
	Issuer              string
	Trusted             error // why browsers would not trust it, or nil
}

// Config says where Cahut is installed.
type Config struct {
	BinPath   string
	DataDir   string
	StateDir  string
	HealthURL string
	Version   string // of this binary
}

type doctor struct {
	ctx     context.Context
	sys     System
	cfg     Config
	results []Result
	health  *update.Health
}

// Run runs every check, in order.
func Run(ctx context.Context, sys System, cfg Config) []Result {
	d := &doctor{ctx: ctx, sys: sys, cfg: cfg}
	d.system()
	d.clock()
	d.service()
	d.data()
	d.ports()
	d.firewall()
	d.selinux()
	d.proxies()
	d.cahuts()
	d.outside()
	d.updates()
	d.add(Info, "Hosting provider", "Ports 80 and 443 must be open in the firewall of your hosting provider too, if it has one.")
	return d.results
}

func (d *doctor) add(s Status, topic, format string, args ...any) {
	d.results = append(d.results, Result{Status: s, Topic: topic, Text: fmt.Sprintf(format, args...)})
}

// Supported systems, by the ID and the major VERSION_ID of os-release.
var supported = map[string][]string{"almalinux": {"9", "10"}, "ubuntu": {"22.04", "24.04"}}

func (d *doctor) system() {
	data, err := d.sys.ReadFile("/etc/os-release")
	if err != nil {
		d.add(Warn, "System", "Cannot read /etc/os-release: Cahut is tested on AlmaLinux 9 and 10, and Ubuntu 22.04 and 24.04.")
		return
	}
	os := map[string]string{}
	lines := bufio.NewScanner(bytes.NewReader(data))
	for lines.Scan() {
		if k, v, ok := strings.Cut(lines.Text(), "="); ok {
			os[k] = strings.Trim(v, `"`)
		}
	}
	name := os["PRETTY_NAME"]
	if name == "" {
		name = os["ID"] + " " + os["VERSION_ID"]
	}
	major := os["VERSION_ID"]
	if os["ID"] != "ubuntu" {
		major, _, _ = strings.Cut(major, ".")
	}
	if slices.Contains(supported[os["ID"]], major) {
		d.add(OK, "System", "%s, %s.", name, runtime.GOARCH)
		return
	}
	d.add(Warn, "System", "%s, %s: Cahut is tested on AlmaLinux 9 and 10, and Ubuntu 22.04 and 24.04. It should work, untested.", name, runtime.GOARCH)
}

func (d *doctor) clock() {
	out, err := d.sys.Command(d.ctx, "timedatectl", "show", "--property=NTPSynchronized", "--value")
	switch {
	case err != nil:
		d.add(Warn, "Clock", "Cannot tell whether the clock is synchronized: certificates need the right time.")
	case out == "yes":
		d.add(OK, "Clock", "Synchronized.")
	default:
		d.add(Fail, "Clock", "Not synchronized: certificates need the right time. Turn synchronization on: sudo timedatectl set-ntp true")
	}
}

func (d *doctor) service() {
	active, _ := d.sys.Command(d.ctx, "systemctl", "is-active", "cahut")
	enabled, _ := d.sys.Command(d.ctx, "systemctl", "is-enabled", "cahut")
	switch {
	case active == "active" && enabled == "enabled":
		d.add(OK, "Service", "The cahut service runs, and starts at boot.")
	case active == "active":
		d.add(Warn, "Service", "The cahut service runs, but does not start at boot: sudo systemctl enable cahut")
	case active == "" && enabled == "":
		d.add(Fail, "Service", "No cahut service: install Cahut with install.sh.")
		return
	default:
		d.add(Fail, "Service", "The cahut service is %s: start it with sudo systemctl start cahut, and journalctl -u cahut says why it stopped.", active)
	}

	h, err := d.readHealth()
	if err != nil {
		d.add(Fail, "Service", "The service does not answer on port 80 of this machine (%v): journalctl -u cahut says why.", err)
		return
	}
	d.health = &h
	if d.cfg.Version != "" && h.Version != d.cfg.Version {
		d.add(Warn, "Service", "The service runs Cahut %s, and the binary is %s: restart it with sudo systemctl restart cahut", h.Version, d.cfg.Version)
		return
	}
	d.add(OK, "Service", "Cahut %s answers, with %s.", h.Version, plural(len(h.Cahuts), "Cahut", "Cahuts"))
}

func (d *doctor) readHealth() (update.Health, error) {
	_, body, err := d.sys.Get(d.ctx, d.cfg.HealthURL)
	if err != nil {
		return update.Health{}, err
	}
	var h update.Health
	if err := json.Unmarshal(body, &h); err != nil {
		return update.Health{}, fmt.Errorf("an answer that is not the health of Cahut: %w", err)
	}
	return h, nil
}

const gib = 1 << 30

func (d *doctor) data() {
	dir := d.cfg.DataDir
	if _, err := d.sys.Stat(dir); err != nil {
		d.add(Fail, "Data", "No data directory %s: install Cahut with install.sh.", dir)
		return
	}
	if owner, err := d.sys.Owner(dir); err == nil && owner != "cahut" {
		d.add(Fail, "Data", "%s belongs to %s, not to the cahut user: sudo chown -R cahut:cahut %s", dir, owner, dir)
	}
	free, err := d.sys.FreeBytes(dir)
	switch {
	case err != nil:
		d.add(Warn, "Data", "Cannot tell the free space of %s.", dir)
	case free < gib:
		d.add(Fail, "Data", "%s has %s free: Windows can no longer be published. Free some space.", dir, size(free))
	case free < 5*gib:
		d.add(Warn, "Data", "%s has %s free.", dir, size(free))
	default:
		d.add(OK, "Data", "%s, %s free.", dir, size(free))
	}
}

func (d *doctor) ports() {
	var closed []string
	for _, port := range []string{"80", "443"} {
		if err := d.sys.Dial(d.ctx, net.JoinHostPort("127.0.0.1", port)); err != nil {
			closed = append(closed, port)
		}
	}
	if len(closed) == 0 {
		d.add(OK, "Ports", "The service listens on ports 80 and 443.")
		return
	}
	d.add(Fail, "Ports", "Nothing listens on port %s of this machine: is the service running, or does another program hold the port? sudo ss -ltnp", strings.Join(closed, " and "))
}

// ufwAllows finds the rules of ufw that let in port 80 or 443.
var ufwAllows = regexp.MustCompile(`(?m)^(80|443)(/tcp)?\s+(\(v6\)\s+)?ALLOW`)

func (d *doctor) firewall() {
	if !d.sys.Root() {
		d.add(Info, "Firewall", "Run cahut doctor as root, with sudo, to check the firewall and SELinux too.")
		return
	}
	if state, _ := d.sys.Command(d.ctx, "systemctl", "is-active", "firewalld"); state == "active" {
		var missing []string
		for _, svc := range []string{"http", "https"} {
			if out, _ := d.sys.Command(d.ctx, "firewall-cmd", "--query-service="+svc); out != "yes" {
				missing = append(missing, svc)
			}
		}
		if len(missing) == 0 {
			d.add(OK, "Firewall", "firewalld lets in ports 80 and 443.")
			return
		}
		d.add(Fail, "Firewall", "firewalld does not let in %s: sudo firewall-cmd --permanent --add-service=http --add-service=https && sudo firewall-cmd --reload",
			strings.Join(missing, " and "))
		return
	}
	if out, err := d.sys.Command(d.ctx, "ufw", "status"); err == nil && strings.Contains(out, "Status: active") {
		allowed := map[string]bool{}
		for _, m := range ufwAllows.FindAllStringSubmatch(out, -1) {
			allowed[m[1]] = true
		}
		if allowed["80"] && allowed["443"] {
			d.add(OK, "Firewall", "ufw lets in ports 80 and 443.")
			return
		}
		d.add(Fail, "Firewall", "ufw does not let in ports 80 and 443: sudo ufw allow 80/tcp && sudo ufw allow 443/tcp")
		return
	}
	d.add(OK, "Firewall", "No firewall runs on this machine.")
}

func (d *doctor) selinux() {
	if !d.sys.Root() {
		return
	}
	mode, err := d.sys.Command(d.ctx, "getenforce")
	if err != nil || mode != "Enforcing" {
		if err == nil {
			d.add(OK, "SELinux", "%s.", mode)
		}
		return
	}
	label, err := d.sys.Label(d.cfg.BinPath)
	if err != nil || !strings.Contains(label, ":bin_t:") {
		d.add(Fail, "SELinux", "SELinux enforces, and %s is labeled %q, not bin_t: systemd will not run it. Fix the label: sudo restorecon %s",
			d.cfg.BinPath, label, d.cfg.BinPath)
		return
	}
	d.add(OK, "SELinux", "Enforcing, and the binary has its label.")
}

// proxies says which proxies the service trusts to pass the address of each
// client, and whether the last connection of one came without a valid
// header. See docs/adr/0041-the-proxy-protocol-from-trusted-proxies.md.
func (d *doctor) proxies() {
	if d.health == nil {
		return
	}
	for _, p := range d.health.ProxyProtocol {
		listener, trusted := strings.ToUpper(p.Listener), strings.Join(p.Trusted, ", ")
		switch {
		case p.LastRefused != nil && p.LastTaken == nil:
			d.add(Fail, "Proxy", "%s on %s has refused every connection of %s since the service started, the last on %s: %s. "+
				"The proxy must send the PROXY protocol, as proxy_protocol on does in nginx, or CAHUT_%s_PROXY_PROTOCOL must be empty.",
				listener, p.Address, trusted, at(*p.LastRefused), p.Refusal, listener)
		case p.LastRefused != nil && p.LastRefused.After(*p.LastTaken):
			d.add(Warn, "Proxy", "%s on %s refused the last connection of %s, on %s: %s. Does every connection of the proxy send the PROXY protocol?",
				listener, p.Address, trusted, at(*p.LastRefused), p.Refusal)
		default:
			d.add(OK, "Proxy", "%s on %s reads the address of each client from %s, by the PROXY protocol.", listener, p.Address, trusted)
		}
	}
}

func (d *doctor) cahuts() {
	if d.health == nil {
		d.add(Warn, "Cahuts", "The service does not answer, so the Cahuts cannot be checked.")
		return
	}
	if len(d.health.Cahuts) == 0 {
		d.add(Fail, "Cahuts", "This server holds no Cahut yet: run install.sh with --domain, or sudo -u cahut %s domain add <domain>", d.cfg.BinPath)
		return
	}
	local, _ := d.sys.LocalAddrs()
	for _, c := range d.health.Cahuts {
		topic := c.Domain
		if !c.OK {
			d.add(Fail, topic, "Does not open: %s. journalctl -u cahut says more.", c.Error)
			continue
		}
		if c.Domain == "localhost" || strings.HasSuffix(c.Domain, ".localhost") {
			d.add(Info, topic, "A Cahut for development, served by cahut serve --dev only.")
			continue
		}
		d.address(topic, c.Domain, local)
		d.certificate(topic, c.Domain)
		d.reach(topic, c.Domain)
	}
}

// address checks that the domain of a Cahut points to this machine.
func (d *doctor) address(topic, domain string, local []net.IP) {
	if net.ParseIP(domain) != nil {
		d.add(OK, topic, "A Cahut at the IP address of the server.")
		return
	}
	ips, err := d.sys.LookupIP(d.ctx, domain)
	if err != nil || len(ips) == 0 {
		d.add(Fail, topic, "%s does not resolve: add an A record (or AAAA) for it, pointing to this server.", domain)
		return
	}
	var names []string
	here, loopback := false, false
	for _, ip := range ips {
		names = append(names, ip.String())
		loopback = loopback || ip.IsLoopback()
		here = here || !ip.IsLoopback() && slices.ContainsFunc(local, ip.Equal)
	}
	if loopback && !here {
		d.add(Fail, topic, "Points to %s, which only this machine reaches: give it an A record (or AAAA) with the public address of the server.", strings.Join(names, ", "))
		return
	}
	if here {
		d.add(OK, topic, "Points to this machine (%s).", strings.Join(names, ", "))
		return
	}
	d.add(Warn, topic, "Points to %s, which is not an address of this machine. Behind NAT, that is fine; otherwise, change its DNS record.", strings.Join(names, ", "))
}

// certificate checks the certificate that the service serves for domain.
func (d *doctor) certificate(topic, domain string) {
	cert, err := d.sys.Certificate(d.ctx, "127.0.0.1:443", domain)
	if err != nil {
		d.add(Fail, topic, "No certificate yet (%v): Let's Encrypt must reach port 80 of %s, from the Internet.", err, domain)
		return
	}
	now := d.sys.Now()
	left := cert.NotAfter.Sub(now)
	switch {
	case cert.Trusted != nil:
		d.add(Warn, topic, "Its certificate, from %s, is not one that browsers trust: %v. Is it from the staging environment of Let's Encrypt?", cert.Issuer, cert.Trusted)
	case left <= 0:
		d.add(Fail, topic, "Its certificate expired on %s: Let's Encrypt must reach port 80 of %s to renew it.", day(cert.NotAfter), domain)
	case left < cert.NotAfter.Sub(cert.NotBefore)/4:
		d.add(Warn, topic, "Its certificate expires on %s, and should have been renewed by now: Let's Encrypt must reach port 80 of %s.", day(cert.NotAfter), domain)
	default:
		d.add(OK, topic, "Its certificate, from %s, is valid until %s.", cert.Issuer, day(cert.NotAfter))
	}
}

// reach asks the Cahut for its health at its own address, as the world
// does, which tries the DNS, the port, the firewalls and the certificate.
func (d *doctor) reach(topic, domain string) {
	host := domain
	if ip := net.ParseIP(domain); ip != nil && ip.To4() == nil {
		host = "[" + domain + "]"
	}
	status, body, err := d.sys.Get(d.ctx, "https://"+host+"/-/healthz")
	if err == nil && status == http.StatusOK && strings.TrimSpace(string(body)) == "ok" {
		d.add(OK, topic, "https://%s/ answers.", host)
		return
	}
	if err == nil {
		err = fmt.Errorf("status %d", status)
	}
	d.add(Warn, topic, "https://%s/ does not answer from this machine (%v). Check its DNS record and the firewall of your hosting provider.", host, err)
}

func (d *doctor) outside() {
	if status, _, err := d.sys.Get(d.ctx, "https://acme-v02.api.letsencrypt.org/directory"); err != nil || status != http.StatusOK {
		d.add(Fail, "Let's Encrypt", "Cannot be reached from this machine: certificates can neither be obtained nor renewed. Check the outgoing traffic of the server.")
	} else {
		d.add(OK, "Let's Encrypt", "Answers.")
	}
	if _, _, err := d.sys.Get(d.ctx, "https://api.github.com/"); err != nil {
		d.add(Warn, "GitHub", "Cannot be reached from this machine: the updater cannot read new releases.")
	} else {
		d.add(OK, "GitHub", "Answers.")
	}
}

func (d *doctor) updates() {
	st, managed, err := update.ReadStatus(d.cfg.StateDir)
	if err != nil {
		d.add(Warn, "Updates", "Cannot read the state of updates: %v", err)
		return
	}
	if !managed {
		d.add(Warn, "Updates", "This server does not update itself: run install.sh again to set up the updater.")
		return
	}
	if state, _ := d.sys.Command(d.ctx, "systemctl", "is-active", "cahut-update.timer"); state != "active" {
		d.add(Warn, "Updates", "The daily update does not run: sudo systemctl enable --now cahut-update.timer")
	} else {
		policy := st.Policy
		if policy == "" {
			policy = update.Patch
		}
		d.add(OK, "Updates", "The daily update runs, with CAHUT_UPDATES=%s.", policy)
	}
	if run := st.Last; run != nil {
		switch run.Result {
		case update.Updated:
			d.add(OK, "Updates", "Updated from %s to %s on %s.", run.From, run.To, day(run.Ended))
		case update.RolledBack:
			d.add(Warn, "Updates", "The update to %s on %s was rolled back: %s", run.To, day(run.Ended), run.Reason)
		case update.Failed:
			d.add(Fail, "Updates", "The update to %s on %s failed, and so did going back: %s", run.To, day(run.Ended), run.Reason)
		default:
			d.add(Warn, "Updates", "The update to %s on %s installed nothing: %s", run.To, day(run.Ended), run.Reason)
		}
	}
	if c := st.Check; c != nil {
		running, rerr := release.ParseVersion(d.cfg.Version)
		latest, lerr := release.ParseVersion(c.Latest)
		switch {
		case c.Error != "":
			d.add(Warn, "Updates", "The last check, on %s, could not read the releases: %s", day(c.At), c.Error)
		case rerr == nil && lerr == nil && latest.Compare(running) > 0:
			d.add(Info, "Updates", "Cahut %s is available: sudo %s update", latest, d.cfg.BinPath)
		}
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

func size(b uint64) string {
	return fmt.Sprintf("%.1f GiB", float64(b)/gib)
}

func day(t time.Time) string {
	return t.UTC().Format("2 January 2006")
}

func at(t time.Time) string {
	return t.UTC().Format("2 January 2006 at 15:04 UTC")
}
